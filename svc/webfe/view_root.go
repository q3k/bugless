// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func (f *httpFrontend) viewRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		q := r.URL.Query().Get("q")
		http.Redirect(w, r, fmt.Sprintf("/issues?q=%s", url.QueryEscape(q)), 302)
		return
	}

	http.NotFound(w, r)
}
