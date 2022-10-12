package auth

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/jlewi/flaap/go/pkg/networking"
	"github.com/kubeflow/internal-acls/google_groups/pkg/gcp/gcs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OIDCWebFlowHelper helps get an OIDC token using the web flow.
// GetTokenSource returns a token source which surfaces the OIDC token as the AccessToken.
// This tokensource can used for Authorization flows that use OIDC tokens as the bearer token.
//
// This flow is useful when obtaining OIDC tokens for human based accounts as these require the user to go through
// an OAuth web flow to generate the credentials.
//
// For robot accounts it should be possible to generate the OIDC token without going through the WebFlow; e.g. by
// using the private key for the robot account.
// See for example: https://pkg.go.dev/google.golang.org/api/idtoken
type OIDCWebFlowHelper struct {
	config *oauth2.Config
	log    logr.Logger
	s      *OIDCWebFlowServer
	ts     oauth2.TokenSource
}

// NewOIDCWebFlowHelper constructs a new web flow helper. oAuthClientFile should be the path to a credentials.json
// downloaded from the API console.
func NewOIDCWebFlowHelper(oAuthClientFile string, issuer string) (*OIDCWebFlowHelper, error) {
	var fHelper gcs.FileHelper

	if strings.HasPrefix(oAuthClientFile, "gs://") {
		ctx := context.Background()
		client, err := storage.NewClient(ctx)

		if err != nil {
			return nil, err
		}

		fHelper = &gcs.GcsHelper{
			Ctx:    ctx,
			Client: client,
		}
	} else {
		fHelper = &gcs.LocalFileHelper{}
	}

	reader, err := fHelper.NewReader(oAuthClientFile)

	if err != nil {
		return nil, err

	}
	b, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}
	// If modifying these scopes, delete your previously saved token.json.
	scopes := []string{oidc.ScopeOpenID, "profile", "email"}
	// "openid" is a required scope for OpenID Connect flows.
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse client secret file to config")
	}

	// TODO(jeremy): make this a parameter. 0 picks a free port.
	port, err := networking.GetFreePort()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get free port")
	}
	config.RedirectURL = fmt.Sprintf("http://127.0.0.1:%v/%v", port, authCallbackUrl)
	p, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create OIDC provider for %v", issuer)
	}

	// Configure an OpenID Connect aware OAuth2 client.
	config.Endpoint = p.Endpoint()

	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}
	verifier := p.Verifier(oidcConfig)

	log := zapr.NewLogger(zap.L())

	s, err := NewOIDCWebFlowServer(*config, verifier, log)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create server")
	}

	// Get a token source
	ts, err := s.Run()

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create OIDC tokensource")
	}

	return &OIDCWebFlowHelper{
		config: config,
		log:    log,
		s:      s,
		ts:     ts,
	}, nil
}

func (h *OIDCWebFlowHelper) GetOAuthConfig() *oauth2.Config {
	return h.config
}

// GetTokenSource requests a token from the web, then returns the retrieved token.
func (h *OIDCWebFlowHelper) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return h.ts, nil
}

// IDTokenSource is a wrapper around a TokenSource that returns the OpenID token as the access token.
type IDTokenSource struct {
	Source   oauth2.TokenSource
	Verifier *oidc.IDTokenVerifier
}

func (s *IDTokenSource) Token() (*oauth2.Token, error) {
	tk, err := s.Source.Token()
	if err != nil {
		return nil, errors.Wrapf(err, "IDTokenSource failed to get token from underlying token source")
	}
	// Per https://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation
	// the ID token is returned in the id_token field
	rawTok, ok := tk.Extra("id_token").(string)
	if !ok {
		return nil, errors.Errorf("Underlying token source didn't have field id_token")
	}

	tok, err := s.Verifier.Verify(context.Background(), rawTok)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to verify ID token")
	}

	// Create a new OAuthToken in which the Access token is the JWT (i.e. the ID token).
	jwtToken := &oauth2.Token{
		AccessToken: rawTok,
		Expiry:      tok.Expiry,
	}

	return jwtToken, nil
}
