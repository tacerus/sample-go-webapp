package web

import (
	"context"
	"encoding/gob"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/tacerus/sample-go-webapp/core"
)

type WebApp struct {
	Oauth2Config   oauth2.Config
	OidcConfig     *oidc.Config
	OidcProvider   *oidc.Provider
	Ctx            context.Context
	Verifier       *oidc.IDTokenVerifier
	sessionManager *scs.SessionManager
	bind           string
	templates      *Templates
	staticDir      string
}

func NewWebApp(c core.Config) *WebApp {
	if c.AssetDir == "" {
		slog.Error("Missing AssetDir.")
		return nil
	}
	if c.Bind == "" {
		slog.Error("Missing Bind.")
		return nil
	}

	app := new(WebApp)
	app.Ctx = context.Background()

	slog.Debug("Initializing OIDC provider ...")

	provider, err := oidc.NewProvider(app.Ctx, c.OidcBaseUrl)
	if err != nil {
		panic(err)
	}

	app.OidcProvider = provider

	app.OidcConfig = &oidc.Config{
		ClientID: c.ClientId,
	}

	app.Verifier = provider.Verifier(app.OidcConfig)

	app.bind = c.Bind

	slog.Debug("Initializing OAuth2 ...")

	bu := c.BaseUrl
	if bu == "" {
		bu = "http://" + app.bind
	}

	app.Oauth2Config = oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  bu + callbackPath,
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
		},
	}

	//app.sessionManager = newSessionManager()

	gob.Register(&oauth2.Token{})

	app.templates = readTemplates(c.AssetDir)
	app.staticDir = c.AssetDir + "/static"

	return app
}

func (app *WebApp) Start() *http.Server {
	app.sessionManager = newSessionManager()

	mux := app.newMux()
	srv := &http.Server{
		Addr:    app.bind,
		Handler: app.sessionManager.LoadAndSave(app.initHandler(mux)),
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	slog.Info("Listening ...", "bind", app.bind)

	return srv
}
