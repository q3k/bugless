package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"code.hackerspace.pl/hscloud/go/mirko"
	cpb "github.com/q3k/bugless/proto/common"
	spb "github.com/q3k/bugless/proto/svc"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"
)

const (
	cookieSession = "bugless_session"
)

// setCookie sets an encrypted, authenticated cookie on the given response.
func (f *httpFrontend) setCookie(w http.ResponseWriter, name string, value []byte) error {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("could not generate nonce: %v", err)
	}

	// Prefix value with cookie name, to prevent using cookie A as cookie B.
	value = append([]byte(name+"\n"), value...)

	var key [32]byte
	copy(key[:], f.secretKey[:32])
	encrypted := secretbox.Seal(nonce[:], value, &nonce, &key)

	// TODO(q3k): set Domain, Secure and SameSite based on app configuration.
	c := &http.Cookie{
		Name:  name,
		Value: base64.StdEncoding.EncodeToString(encrypted),
		Path:  "/",
	}
	http.SetCookie(w, c)
	return nil
}

func (f *httpFrontend) getCookie(r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, fmt.Errorf("retrieving cookie: %w", err)
	}
	if len(cookie.Value) > 1024 || len(cookie.Value) == 0 {
		return nil, fmt.Errorf("cookie invalid: no value, or value too long")
	}

	encrypted, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("cookie invalid: could not decode base64")
	}

	// Extract nonce from encrypted payload (first 24 bytes)
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])

	var key [32]byte
	copy(key[:], f.secretKey[:32])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("cookie invalid: decryption failed")
	}

	// Check and remove cookie name prefix.
	if !bytes.HasPrefix(decrypted, []byte(name+"\n")) {
		return nil, fmt.Errorf("cookie invalid: invalid cookie name prefix")
	}
	return decrypted[len(name)+1:], nil
}

type session struct {
	accessToken            string
	tokenExpiration        time.Time
	username               string
	denormalizedExpiration time.Time
}

func (s *session) token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: s.accessToken,
	}
}

func (s *session) proto() *spb.WebSession {
	return &spb.WebSession{
		AccessToken:            s.accessToken,
		TokenExpiration:        &cpb.Timestamp{Nanos: s.tokenExpiration.UnixNano()},
		Username:               s.username,
		DenormalizedExpiration: &cpb.Timestamp{Nanos: s.denormalizedExpiration.UnixNano()},
	}
}

func (s *session) soy() map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"username": s.username,
	}
}

// check ensures that the session is valid, ie.:
// - the access token is not expired
// - if the denormalize data expired, that is still valid, and
//   that the access token has not been revoked
// It will return an updated session on success, and nil if the session was
// deemed to not be valid.
func (s *session) check(ctx context.Context, f *httpFrontend) *session {
	now := time.Now()

	// Did the access token expire? Expire the session fully.
	if now.After(s.tokenExpiration) {
		return nil
	}

	sess := *s

	// Did the denormalized data expire? Get it again, and update the session.
	if now.After(s.denormalizedExpiration) {
		mirko.TraceInfof(ctx, "refreshing denormalized session data...")
		ui, err := f.oidc.UserInfo(ctx, oauth2.StaticTokenSource(s.token()))
		if err != nil {
			mirko.TraceErrorf(ctx, "failed to get UserInfo on denormalized data refresh: %v", err)
			return nil
		}
		sess.username = ui.Email
		sess.denormalizedExpiration = time.Now().Add(time.Hour)
	}

	return &sess
}

func (f *httpFrontend) getSession(w http.ResponseWriter, r *http.Request) *session {
	ctx := r.Context()

	// We swallow any cookie retrieval errors into 'no session'. This makes
	// sense for a flow where "you're not logged in" is the proper way to fail
	// safely.

	b, err := f.getCookie(r, cookieSession)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			f.setSession(w, nil)
			mirko.TraceErrorf(ctx, "could not retrieve session cookie: %v", err)
		}
		return nil
	}

	var sessionProto spb.WebSession
	if err := proto.Unmarshal(b, &sessionProto); err != nil {
		f.setSession(w, nil)
		mirko.TraceErrorf(ctx, "could not unmarshal session cookie: %v", err)
		return nil
	}

	s := &session{
		accessToken:            sessionProto.AccessToken,
		tokenExpiration:        time.Unix(0, sessionProto.TokenExpiration.Nanos),
		username:               sessionProto.Username,
		denormalizedExpiration: time.Unix(0, sessionProto.DenormalizedExpiration.Nanos),
	}

	s = s.check(ctx, f)
	if err := f.setSession(w, s); err != nil {
		mirko.TraceErrorf(r.Context(), "could not save checked session after retrieval: %v", err)
		return nil
	}

	return s
}

// setSession sets a session cookie, or if nil, remove it.
func (f *httpFrontend) setSession(w http.ResponseWriter, s *session) error {
	if s == nil {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieSession,
			MaxAge: -1,
		})
		return nil
	}
	b, err := proto.Marshal(s.proto())
	if err != nil {
		return fmt.Errorf("marshaling session cookie: %w", err)
	}

	return f.setCookie(w, cookieSession, b)
}
