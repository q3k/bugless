package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"code.hackerspace.pl/hscloud/go/mirko"

	"golang.org/x/oauth2"
)

const (
	cookieOAuthState = "bugless_oauth_state"
)

func (f *httpFrontend) viewLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Generate anti-CSRF 'state' value, as per RFC 6749 Section 10.12.
	// On an incoming redirection URL call, this state will be verified to be
	// the same as generated here. This value is kept in an encrypted HTTP
	// cookie sent to the client alongside redirecting it to the authorization
	// URI.
	var stateBuf [8]byte
	if _, err := rand.Read(stateBuf[:]); err != nil {
		mirko.TraceErrorf(ctx, "could not generate anti-CSRF state: %v", err)
		f.internalError(w)
		return
	}
	state := hex.EncodeToString(stateBuf[:])

	// Save this state to an encrypted, authenticated cookie.
	if err := f.setCookie(w, cookieOAuthState, []byte(state)); err != nil {
		mirko.TraceErrorf(ctx, "could not save state cookie: %v", err)
		f.internalError(w)
		return
	}

	// Redirect user to authorization endpoint.
	url := f.oauth2.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, 302)
}

func (f *httpFrontend) viewLoginOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fail := func() {
		w.WriteHeader(402)
		fmt.Fprintf(w, "Sorry, we couldn't log you in. Try again?")
	}

	// Validate state against cookie.
	stateCookie, err := f.getCookie(r, cookieOAuthState)
	if err != nil {
		mirko.TraceErrorf(ctx, "could not retrieve state cookie: %v", err)
		fail()
		return
	}

	state := r.FormValue("state")
	if state != string(stateCookie) {
		mirko.TraceErrorf(ctx, "state mismatch. cookie %q, redirect %q", string(stateCookie), state)
		fail()
		return
	}

	token, err := f.oauth2.Exchange(ctx, r.FormValue("code"), oauth2.AccessTypeOnline)
	if err != nil {
		mirko.TraceErrorf(ctx, "exchange: %v", err)
		fail()
		return
	}

	sess := &session{
		accessToken:     token.AccessToken,
		tokenExpiration: token.Expiry,
	}

	if err := f.setSession(w, sess.check(ctx, f)); err != nil {
		mirko.TraceErrorf(ctx, "setSession: %v", err)
		fail()
		return
	}

	http.Redirect(w, r, "/", 302)
}

func (f *httpFrontend) viewLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := f.setSession(w, nil); err != nil {
		mirko.TraceErrorf(ctx, "setSession: %v", err)
	}

	http.Redirect(w, r, "/", 302)
}
