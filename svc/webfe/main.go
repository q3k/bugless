// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"net/http"

	"github.com/inconshreveable/log15"
	gosoy "github.com/robfig/soy"

	"github.com/q3k/bugless/svc/webfe/gss"
	"github.com/q3k/bugless/svc/webfe/js"
	"github.com/q3k/bugless/svc/webfe/soy"
)

func main() {
	bundle := gosoy.NewBundle()
	for k, v := range soy.Data {
		log15.Info("loading soy template", "name", k)
		bundle = bundle.AddTemplateString(k, string(v))
	}
	tofu, err := bundle.CompileToTofu()
	if err != nil {
		log15.Crit("could not build tofu", "err", err)
		return
	}

	http.HandleFunc("/tofu", func(w http.ResponseWriter, r *http.Request) {
		tofu.Render(w, "bugless.templates.note", map[string]interface{}{
			"title":   "foo",
			"content": "baz",
		})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tofu.Render(w, "bugless.templates.base.html", map[string]interface{}{
			"title": "Bugless - Home",
		})
	})

	http.HandleFunc("/js.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(js.Data["js.js"])
	})
	http.HandleFunc("/gss.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(gss.Data["gss.css"])
	})

	log15.Info("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log15.Error("could not listen", "err", err)
	}
}
