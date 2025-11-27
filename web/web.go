/*
   Copyright (C) 2025  SUSE LLC <georg.pfuetzenreuter@suse.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const callbackPath = "/login/callback"

func (app *WebApp) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.indexHandler)
	mux.HandleFunc("/login/init", app.loginHandler)
	mux.HandleFunc(callbackPath, app.callbackHandler)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(app.staticDir))))

	return mux
}

func (app *WebApp) setOrigPath(r *http.Request) {
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}
	slog.DebugContext(r.Context(), "setting origPath", "path", path)
	app.sessionManager.Put(r.Context(), "origPath", path)
}

func (app *WebApp) getOrigPath(r *http.Request) string {
	path := app.sessionManager.GetString(r.Context(), "origPath")
	if path == "" {
		path = "/"
	}

	return path
}

func (app *WebApp) getSessionId(r *http.Request) string {
	return app.sessionManager.GetString(r.Context(), "id")
}

func (app *WebApp) initHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		had_session_id := true

		session_id := app.getSessionId(r)
		if session_id == "" {
			had_session_id = false

			var err error
			session_id, err = randString(12, false)
			if err != nil {
				slog.ErrorContext(r.Context(), "Session ID generation failed", "error", err)
				app.errorHandler(w, r, ERR_MISC, "")
				return
			}

			// for display on pages
			app.sessionManager.Put(r.Context(), "id", session_id)
		}

		// for logging
		r = r.WithContext(context.WithValue(r.Context(), "session_id", session_id))

		if !had_session_id {
			slog.DebugContext(r.Context(), "Initialized session")
		}

		next.ServeHTTP(w, r)
	})
}

func (app *WebApp) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Add("Content-Type", "text/html")

	subject := app.sessionManager.GetString(r.Context(), "subject")
	p := PageData{
		Subject:   subject,
		SessionId: app.getSessionId(r),
	}
	app.templates.Index.ExecuteTemplate(w, "base", p)
}

func (app *WebApp) getSubject(w http.ResponseWriter, r *http.Request) (string, bool) {
	token := app.sessionManager.Get(r.Context(), "token").(*oauth2.Token)
	if token == nil {
		slog.ErrorContext(r.Context(), "Subject query attempted without token")
		app.errorHandler(w, r, ERR_MISC, "Subject query failed (no token).")
	}

	ui, err := app.OidcProvider.UserInfo(app.Ctx, app.Oauth2Config.TokenSource(app.Ctx, token))
	if err != nil || ui.Subject == "" {
		slog.ErrorContext(r.Context(), "Failed to query userinfo for subject", "error", err)
		app.errorHandler(w, r, ERR_MISC, "Subject query failed (failed to query user info).")

		return "", false
	}

	return ui.Subject, true
}

func (app *WebApp) checkSession(w http.ResponseWriter, r *http.Request) bool {
	subject_session := app.sessionManager.GetString(r.Context(), "subject")

	if subject_session == "" || app.sessionManager.GetString(r.Context(), "state") == "" || app.sessionManager.GetString(r.Context(), "nonce") == "" {
		app.errorHandler(w, r, ERR_STATE, "Empty subject session, state or nonce.")
		return false
	}

	subject_userinfo, ok := app.getSubject(w, r)
	if !ok {
		return false
	}

	if subject_session != subject_userinfo {
		slog.WarnContext(r.Context(), "Privileged action attempted with mismatching subject", "subject_session", subject_session, "subject_userinfo", subject_userinfo)
		app.errorHandler(w, r, ERR_ILLEG, "Subject mismatch.")
		return false
	}

	return true
}

func (app *WebApp) checkState(w http.ResponseWriter, r *http.Request) bool {
	if app.sessionManager.GetString(r.Context(), "state") != r.URL.Query().Get("state") {
		app.errorHandler(w, r, ERR_STATE, "State mismatch.")
		return false
	}

	return true
}

func (app *WebApp) loginHandler(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16, true)
	if err != nil {
		slog.ErrorContext(r.Context(), "randString() failed", "error", err)
		app.errorHandler(w, r, ERR_MISC, "No random string.")
		return
	}
	nonce, err := randString(16, true)
	if err != nil {
		slog.ErrorContext(r.Context(), "randString() failed", "error", err)
		app.errorHandler(w, r, ERR_MISC, "No random string.")
		return
	}

	app.sessionManager.Put(r.Context(), "state", state)
	app.sessionManager.Put(r.Context(), "nonce", nonce)

	http.Redirect(w, r, app.Oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (app *WebApp) callbackHandler(w http.ResponseWriter, r *http.Request) {
	if !app.checkState(w, r) {
		return
	}

	code := r.URL.Query().Get("code")

	if code == "" {
		app.errorHandler(w, r, ERR_CODE, "")
		return
	}

	oauth2Token, err := app.Oauth2Config.Exchange(app.Ctx, code)
	if err != nil {
		slog.ErrorContext(r.Context(), "Token exchange failed", "error", err)
		app.errorHandler(w, r, ERR_MISC, "Token exchange failed.")
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		slog.ErrorContext(r.Context(), "No id_token field in oauth2 token")
		app.errorHandler(w, r, ERR_MISC, "Missing id_token field.")
		return
	}

	idToken, err := app.Verifier.Verify(app.Ctx, rawIDToken)
	if err != nil {
		slog.ErrorContext(r.Context(), "ID token verification failed", "error", err)
		app.errorHandler(w, r, ERR_MISC, "ID token verification failed.")
		return
	}

	nonce := app.sessionManager.GetString(r.Context(), "nonce")
	if idToken.Nonce != nonce {
		slog.ErrorContext(r.Context(), "Nonce does not match")
		app.errorHandler(w, r, ERR_MISC, "Nonce verification failed.")
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to renew session token", "error", err)
		if err := app.sessionManager.Destroy(r.Context()); err != nil {
			slog.ErrorContext(r.Context(), "Failed to destroy session", "error", err)
		}
		app.errorHandler(w, r, ERR_MISC, "Session renewal failed.")
		return
	}

	app.sessionManager.Put(r.Context(), "token", oauth2Token)
	app.sessionManager.Put(r.Context(), "token_used", false)
	subject, ok := app.getSubject(w, r)
	if !ok {
		return
	}

	slog.InfoContext(r.Context(), "Authenticated user", "subject", subject)
	app.sessionManager.Put(r.Context(), "subject", subject)

	if !app.checkSession(w, r) {
		return
	}

	http.Redirect(w, r, app.getOrigPath(r), http.StatusFound)
}
