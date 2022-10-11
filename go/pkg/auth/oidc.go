package auth

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/jlewi/flaap/go/pkg/networking"
	"github.com/kubeflow/internal-acls/google_groups/pkg/gcp/gcs"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"strings"
)

// OIDCWebFlowHelper helps get an OIDC token using the web flow.
type OIDCWebFlowHelper struct {
	config *oauth2.Config
	log    logr.Logger
	s      *Server
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

	s, err := NewServer(*config, verifier, log)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create server")
	}

	// Run the server in a background thread.
	go func() {
		s.StartAndBlock()
	}()

	return &OIDCWebFlowHelper{
		config: config,
		log:    log,
		s:      s,
	}, nil
}

func (h *OIDCWebFlowHelper) GetOAuthConfig() *oauth2.Config {
	return h.config
}

// GetTokenSource requests a token from the web, then returns the retrieved token.
func (h *OIDCWebFlowHelper) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	authURL := h.s.AuthStartURL()
	h.log.Info("Opening URL to start Auth Flow", "URL", authURL)
	if err := browser.OpenURL(authURL); err != nil {
		h.log.Error(err, "Failed to open URL in browser", "url", authURL)
		fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)
	}

	tok, err := h.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to retrieve token from web: %v")
	}

	return h.config.TokenSource(ctx, tok), nil
}
