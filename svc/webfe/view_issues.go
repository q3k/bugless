// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	pb "github.com/q3k/bugless/proto/svc"
)

func (f *httpFrontend) viewIssues(w http.ResponseWriter, r *http.Request) {
	session := f.getSession(w, r)
	f.l.Info("session", "session", session)

	q := r.URL.Query().Get("q")
	if r.URL.Path != "/issues" {
		http.Redirect(w, r, fmt.Sprintf("/issues?q=%s", url.QueryEscape(q)), 302)
		return
	}
	if q == "" && session != nil {
		q = "author:" + session.username
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
	var queryErrors []string
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
			if chunk.QueryErrors != nil {
				queryErrors = append(queryErrors, chunk.QueryErrors...)
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
		"title":       "Bugless - Home",
		"lvr":         f.lvr,
		"query":       q,
		"queryErrors": queryErrors,
		"issues":      issues,
		"session":     session.soy(),
	})
	if err == nil {
		return
	}
	f.l.Crit("could not render template", "err", err)
	fmt.Fprintf(w, "something went wrong.")
}
