package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/jlewi/p22h/backend/api"
	"github.com/jlewi/p22h/backend/pkg/debug"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	authStartPrefix = "/auth/start"
	authCallbackUrl = "/auth/callback"
)

type tokenSourceOrError struct {
	ts  oauth2.TokenSource
	err error
}

// OIDCWebFlowServer creates a server to be used as part of client registration in the OIDC protocol.
//
// It is based on the code in https://github.com/coreos/go-oidc/blob/v3/example/idtoken/app.go.
//
// N.B: https://github.com/coreos/go-oidc/issues/354 is discussing creating a reusable server.
type OIDCWebFlowServer struct {
	log      logr.Logger
	listener net.Listener
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
	mu       sync.Mutex
	tokSrc   oauth2.TokenSource
	host     string
	c        chan tokenSourceOrError
	srv      *http.Server
}

func NewOIDCWebFlowServer(config oauth2.Config, verifier *oidc.IDTokenVerifier, log logr.Logger) (*OIDCWebFlowServer, error) {
	u, err := url.Parse(config.RedirectURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not parse URL %v", config.RedirectURL)
	}

	log.Info("Creating listener", "host", u.Host)
	listener, err := net.Listen("tcp", u.Host)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create listener")
	}

	return &OIDCWebFlowServer{
		log:      log,
		listener: listener,
		config:   config,
		verifier: verifier,
		host:     u.Host,
		c:        make(chan tokenSourceOrError, 10),
	}, nil
}

func (s *OIDCWebFlowServer) Address() string {
	return fmt.Sprintf("http://%v", s.host)
}

// AuthStartURL returns the URL to kickoff the oauth login flow.
func (s *OIDCWebFlowServer) AuthStartURL() string {
	return s.Address() + authStartPrefix
}

func (s *OIDCWebFlowServer) writeStatus(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := api.RequestStatus{
		Kind:    "RequestStatus",
		Message: message,
		Code:    code,
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		s.log.Error(err, "Failed to marshal RequestStatus", "RequestStatus", resp, "code", code)
	}

	if code != http.StatusOK {
		caller := debug.ThisCaller()
		s.log.Info("HTTP error", "RequestStatus", resp, "code", code, "caller", caller)
	}
}

func (s *OIDCWebFlowServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	s.writeStatus(w, "OIDC server is running", http.StatusOK)
}

func (s *OIDCWebFlowServer) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	s.writeStatus(w, fmt.Sprintf("OIDC server doesn't handle the path; url: %v", r.URL), http.StatusNotFound)
}

// waitForReady waits until the server is health.
func (s *OIDCWebFlowServer) waitForReady() error {
	endTime := time.Now().Add(3 * time.Minute)
	for time.Now().Before(endTime) {

		r, err := http.Get(s.Address() + "/healthz")
		if err == nil && r.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("timeout waiting for server to be healthy")
}

// Run runs the flow to create a tokensource.
func (s *OIDCWebFlowServer) Run() (oauth2.TokenSource, error) {
	log := s.log

	go func() {
		s.startAndBlock()
	}()
	log.Info("Waiting for OIDC server to be ready")
	if err := s.waitForReady(); err != nil {
		return nil, err
	}
	authURL := s.AuthStartURL()
	log.Info("Opening URL to start Auth Flow", "URL", authURL)
	if err := browser.OpenURL(authURL); err != nil {
		log.Error(err, "Failed to open URL in browser; open it manually", "url", authURL)
		fmt.Printf("Go to the following link in your browser to complete  the OIDC flow: %v\n", authURL)
	}
	// Wait for the token source
	log.Info("Waiting for OIDC login flow to complete")

	defer func() {
		log.Info("Shutting OIDC server down")
		s.srv.Shutdown(context.Background())
	}()
	select {
	case tsOrError := <-s.c:
		if tsOrError.err != nil {
			return nil, errors.Wrapf(tsOrError.err, "OIDC flow didn't complete successfully")
		}
		log.Info("OIDC flow completed")
		return tsOrError.ts, nil
	case <-time.After(3 * time.Minute):
		return nil, errors.New("Timeout waiting for OIDC flow to complete")
	}
}

// startAndBlock starts the server and blocks.
func (s *OIDCWebFlowServer) startAndBlock() {
	log := s.log

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc(authStartPrefix, s.handleStartWebFlow)
	router.HandleFunc("/healthz", s.HealthCheck)
	router.HandleFunc(authCallbackUrl, s.handleAuthCallback)

	router.NotFoundHandler = http.HandlerFunc(s.NotFoundHandler)

	log.Info("OIDC server is running", "address", s.Address())

	s.srv = &http.Server{Addr: s.host, Handler: router}

	s.srv.ListenAndServe()
	err := http.Serve(s.listener, router)

	if err != nil {
		log.Error(err, "OIDCWebFlowServer returned error")
	}
	log.Info("OIDC server has been shutdown")
}

// handleStartWebFlow kicks off the OIDC web flow.
// It was copied from: https://github.com/coreos/go-oidc/blob/2cafe189143f4a454e8b4087ef892be64b1c77df/example/idtoken/app.go#L65
// It sets some cookies before redirecting to the OIDC provider's URL for obtaining an authorization code.
func (s *OIDCWebFlowServer) handleStartWebFlow(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		s.c <- tokenSourceOrError{err: errors.Wrapf(err, "Failed to generate state")}
		return
	}
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		s.c <- tokenSourceOrError{err: errors.Wrapf(err, "Failed to generate nonce")}
		return
	}

	// TODO(jeremy): Cookies are not port specific;
	// see https://stackoverflow.com/questions/1612177/are-http-cookies-port-specific#:~:text=Cookies%20do%20not%20provide%20isolation%20by%20port.
	// So if we have two completely instances of the OIDCWebFlowServer running (e.g. in different CLIs) corresponding to two
	// different ports  e.g 127.0.0.1:50002 & 127.0.0.1:60090 the would both be reading/writing the same cookies
	// if the user was somehow going simultaneously going through the flow on both browsers. Extremely unlikely
	// but could still cause concurrency issues. We should address that by adding some random salt to each
	// cookie name at server construction.
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	redirectURL := s.config.AuthCodeURL(state, oidc.Nonce(nonce))

	s.log.V(logging.Debug).Info("Setting redirect URL", "state", state, "nonce", nonce, "url", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// handleAuthCallback handles the OIDC auth callback code copied from
// https://github.com/coreos/go-oidc/blob/2cafe189143f4a454e8b4087ef892be64b1c77df/example/idtoken/app.go#L82.
//
// The Auth callback is invoked in step 21 of the OIDC protocol.
// https://solid.github.io/solid-oidc/primer/#:~:text=Solid%2DOIDC%20builds%20on%20top,authentication%20in%20the%20Solid%20ecosystem.
// The OpenID server responds with a 303 redirect to the AuthCallback URL and passes the authorization code.
// This is a mechanism for the authorization code to be passed into the code.
func (s *OIDCWebFlowServer) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		s.c <- tokenSourceOrError{err: errors.New("state cookie not set")}
		return
	}
	actual := r.URL.Query().Get("state")
	if actual != state.Value {
		s.log.Info("state dind't match", "got", actual, "want", state.Value)
		http.Error(w, "state did not match", http.StatusBadRequest)
		s.c <- tokenSourceOrError{err: errors.New("state argument didn't match value in cookie")}
		return
	}

	oauth2Token, err := s.config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		s.c <- tokenSourceOrError{err: errors.Wrapf(err, "Failed to exchange token")}
		return
	}

	// Create a tokensource. This will take care of automatically refreshing the token if necessary
	// Make a copy of oauth2Token since we will modify it below
	copy := *oauth2Token
	ts := s.config.TokenSource(ctx, &copy)

	// Create an ID token source that wraps this token source
	idTS := &IDTokenSource{
		Source:   ts,
		Verifier: s.verifier,
	}

	// We want to emit the tokenSource after the server has served the page because the channel is used to signal
	// that the flow has completed and therefore the server can be shutdown.
	defer func() {
		s.c <- tokenSourceOrError{ts: idTS}
	}()

	// Code writes the JSON version of the token to the web page.
	// TODO(jeremy): Should we return html that says something like; here's your token please return to your
	// application and close the webbrowser
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce did not match", http.StatusBadRequest)
		return
	}

	oauth2Token.AccessToken = "*REDACTED*"
	oauth2Token.RefreshToken = "*REDACTED*"

	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "OIDC Flow completed successfully.\n")
	fmt.Fprintf(w, "Please close the browser and return to your application.\n")
	fmt.Fprintf(w, "Information about the OIDC token is provided below.\n")
	w.Write(data)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// copied from: https://github.com/coreos/go-oidc/blob/2cafe189143f4a454e8b4087ef892be64b1c77df/example/idtoken/app.go#L34
func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		// See: https://medium.com/swlh/7-keys-to-the-mystery-of-a-missing-cookie-fdf22b012f09
		// Match all paths
		Path: "/",
	}
	http.SetCookie(w, c)
}
