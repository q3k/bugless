// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"net/http"

	"code.hackerspace.pl/hscloud/go/mirko"
	"code.hackerspace.pl/hscloud/go/pki"
	pb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/webfe/gss"
	"github.com/q3k/bugless/svc/webfe/js"
	"github.com/q3k/bugless/svc/webfe/soy"

	oidc "github.com/coreos/go-oidc"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/inconshreveable/log15"
	gosoy "github.com/robfig/soy"
	"github.com/robfig/soy/soyhtml"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

var (
	flagModel             string
	flagPublicHTTPAddress string
	flagSecret            string
	flagOIDCProvider      string
	flagOIDCClientID      string
	flagOIDCClientSecret  string
	flagOIDCRedirectURL   string
)

func init() {
	flag.Set("listen_address", "127.0.0.1:4210")
	flag.Set("debug_address", "127.0.0.1:4211")
}

type httpFrontend struct {
	l     log.Logger
	model pb.ModelClient
	// lvr is the ibazel Live Reload script URL.
	lvr string
	// tofu is the soy/tofu html template bundle.
	tofu *soyhtml.Tofu
	// oidc is the oidc provider client
	oidc *oidc.Provider
	// oauth2 is the oauth2 client config
	oauth2 *oauth2.Config
	// secretKey is a secret used to encrypt and authenticate client-side cookies.
	secretKey []byte

	// paths to static content
	paths struct {
		jsMap string
		js    string
		css   string
	}
}

func (f *httpFrontend) internalError(w http.ResponseWriter) {
	w.WriteHeader(500)
	fmt.Fprintf(w, "An internal server error occured. Sorry, please try again later.")
}

func main() {
	flag.StringVar(&flagPublicHTTPAddress, "public_http_address", "127.0.0.1:8080", "Address to listen on for public HTTP connections")
	flag.StringVar(&flagSecret, "secret", "", "Secret used to encrypt sensitive data in user cookies. Must be shared across all frontend instances")
	flag.StringVar(&flagModel, "model", "127.0.0.1:4200", "Address of bugless model service")
	flag.StringVar(&flagOIDCProvider, "oidc_provider", "https://sso.hackerspace.pl", "Address of OpenID Connect provider")
	flag.StringVar(&flagOIDCClientID, "oidc_client_id", "", "OIDC Client ID")
	flag.StringVar(&flagOIDCClientSecret, "oidc_client_secret", "", "OIDC Client Secret")
	flag.Parse()

	m := mirko.New()
	l := log.New()

	if len(flagSecret) < 8 {
		l.Crit("secret must be at least 8 characters long")
		return
	}

	if flagOIDCClientID == "" {
		l.Crit("oidc_client_id must be set")
		return
	}

	if flagOIDCClientSecret == "" {
		l.Crit("oidc_client_secret must be set")
		return
	}

	// Live Reload currently disabled because it seems broken (reloads before HTTP server is up).
	lvr := ""
	//lvr := os.Getenv("IBAZEL_LIVERELOAD_URL")
	//if lvr != "" {
	//	l.Info("live reload enabled", "url", lvr)
	//}

	bundle := gosoy.NewBundle()
	for k, v := range soy.Data {
		l.Info("loading soy template", "name", k)
		bundle = bundle.AddTemplateString(k, string(v))
	}
	tofu, err := bundle.CompileToTofu()
	if err != nil {
		l.Crit("could not build tofu", "err", err)
		return
	}

	provider, err := oidc.NewProvider(m.Context(), flagOIDCProvider)
	if err != nil {
		l.Crit("could not setup oidc provider", "err", err)
		return
	}

	oauth2Config := &oauth2.Config{
		ClientID:     flagOIDCClientID,
		ClientSecret: flagOIDCClientSecret,
		RedirectURL:  flagOIDCRedirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile:read"},
	}

	// The salt isn't important - we're just using PBKDF2 as a glorified
	// string-to-bytes encoding. We rely on the randmoness of the given secret
	// to provide security.
	secretKey := pbkdf2.Key([]byte(flagSecret), []byte("bugless!"), 4096, 32, sha256.New)

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	conn, err := grpc.Dial(flagModel, pki.WithClientHSPKI())
	if err != nil {
		l.Crit("could not dial model", "err", err)
		return
	}

	proxy := &backendProxy{
		model: pb.NewModelClient(conn),
	}

	mux := http.NewServeMux()

	fe := &httpFrontend{
		model:     pb.NewModelClient(conn),
		l:         l.New("service", "frontend"),
		lvr:       lvr,
		tofu:      tofu,
		oidc:      provider,
		oauth2:    oauth2Config,
		secretKey: secretKey,
	}

	grpcWebServer := grpc.NewServer()
	pb.RegisterModelServer(grpcWebServer, proxy)
	wrappedGrpc := grpcweb.WrapServer(grpcWebServer)

	mux.Handle("/rpc/", http.StripPrefix("/rpc", http.HandlerFunc(wrappedGrpc.ServeHTTP)))

	mux.HandleFunc("/", fe.viewIssues)
	mux.HandleFunc("/issues", fe.viewIssues)
	mux.HandleFunc("/login", fe.viewLogin)
	mux.HandleFunc("/login/oauth-redirect", fe.viewLoginOAuthRedirect)
	mux.HandleFunc("/logout", fe.viewLogout)

	fe.paths.jsMap = serveHashedStatic(mux, "js.js.map", "application/json", js.Data["js.js.map"])
	jsData := []byte(fmt.Sprintf("//# sourceMappingURL=%s\n", fe.paths.jsMap))
	jsData = append(jsData, js.Data["js.js"]...)
	fe.paths.js = serveHashedStatic(mux, "js.js", "text/javascript", jsData)
	fe.paths.css = serveHashedStatic(mux, "gss.css", "text/css", gss.Data["gss.css"])

	ctx := m.Context()
	server := &http.Server{Addr: flagPublicHTTPAddress, Handler: mux}
	go func() {
		l.Info("Listening for public HTTP", "address", flagPublicHTTPAddress)
		err := server.ListenAndServe()
		if err != nil && err != ctx.Err() {
			l.Crit("ListenAndServe: %v", err)
		}
	}()

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}
