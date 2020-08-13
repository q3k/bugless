// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"code.hackerspace.pl/hscloud/go/mirko"
	"code.hackerspace.pl/hscloud/go/pki"
	pb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/webfe/gss"
	"github.com/q3k/bugless/svc/webfe/js"
	"github.com/q3k/bugless/svc/webfe/soy"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/inconshreveable/log15"
	gosoy "github.com/robfig/soy"
	"github.com/robfig/soy/soyhtml"
	"google.golang.org/grpc"
)

var (
	flagModel             string
	flagPublicHTTPAddress string
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
}

func (f *httpFrontend) viewIssues(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if r.URL.Path != "/issues" {
		http.Redirect(w, r, fmt.Sprintf("/issues?q=%s", url.QueryEscape(q)), 302)
		return
	}
	if q == "" {
		// TODO(q3k): unhardcode this once we get authn & user storage
		q = "author:q3k@q3k.org"
		http.Redirect(w, r, fmt.Sprintf("/issues?q=%s", url.QueryEscape(q)), 302)
		return
	}

	stream, err := f.model.GetIssues(r.Context(), &pb.ModelGetIssuesRequest{
		Query: &pb.ModelGetIssuesRequest_BySearch_{
			BySearch: &pb.ModelGetIssuesRequest_BySearch{
				Search: q,
			},
		},
		OrderBy: pb.ModelGetIssuesRequest_ORDER_BY_LAST_UPDATE,
	})

	var issues []map[string]interface{}
	var issuesGetErr error
	if err != nil {
		issuesGetErr = err
	}

	if err == nil {
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				issuesGetErr = err
				break
			}
			for _, issue := range chunk.Issues {
				issues = append(issues, map[string]interface{}{
					"priority":     fmt.Sprintf("%d", issue.Current.Priority),
					"id":           fmt.Sprintf("%d", issue.Id),
					"type":         issueTypePretty(issue.Current.Type),
					"title":        issue.Current.Title,
					"assignee":     issue.Current.Assignee.Id,
					"status":       issueStatusPretty(issue.Current.Status),
					"last_updated": time.Unix(0, issue.LastUpdated.Nanos).Format("Jan 2, 2006 15:04:05"),
				})
			}
		}
	}

	if issuesGetErr != nil {
		// TODO: expose to HTML
		f.l.Error("could not get issues", "err", issuesGetErr)
	}

	err = f.tofu.Render(w, "bugless.templates.base.html", map[string]interface{}{
		"title":  "Bugless - Home",
		"lvr":    f.lvr,
		"query":  q,
		"issues": issues,
	})
	if err == nil {
		return
	}
	f.l.Crit("could not render template", "err", err)
	fmt.Fprintf(w, "something went wrong.")
}

func main() {
	flag.StringVar(&flagPublicHTTPAddress, "public_http_address", "127.0.0.1:8080", "Address to listen on for public HTTP connections")
	flag.StringVar(&flagModel, "model", "127.0.0.1:4200", "Address of bugless model service")
	flag.Parse()
	m := mirko.New()
	l := log.New()

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
	fe := &httpFrontend{
		model: pb.NewModelClient(conn),
		l:     l.New("service", "frontend"),
		lvr:   lvr,
		tofu:  tofu,
	}

	grpcWebServer := grpc.NewServer()
	pb.RegisterModelServer(grpcWebServer, proxy)
	wrappedGrpc := grpcweb.WrapServer(grpcWebServer)

	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", http.HandlerFunc(wrappedGrpc.ServeHTTP)))

	mux.HandleFunc("/", fe.viewIssues)
	mux.HandleFunc("/issues", fe.viewIssues)

	mux.HandleFunc("/js.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		fmt.Fprintf(w, "//# sourceMappingURL=/js.js.map\n")
		w.Write(js.Data["js.js"])
	})
	mux.HandleFunc("/js.js.map", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js.Data["js.js.map"])
	})
	mux.HandleFunc("/gss.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(gss.Data["gss.css"])
	})

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
