// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type issue struct {
	Title string
}

func (i *issue) Type() string {
	return "issue"
}

type update struct {
	Comment string
}

func (u *update) Type() string {
	return "update"
}

func createMapping() *mapping.IndexMappingImpl {
	mapping := bleve.NewIndexMapping()

	issueMapping := bleve.NewDocumentMapping()
	titleFieldNameMapping := bleve.NewTextFieldMapping()
	titleFieldNameMapping.Name = "title"
	issueMapping.AddFieldMappingsAt("title", titleFieldNameMapping)
	mapping.AddDocumentMapping("issue", issueMapping)

	updateMapping := bleve.NewDocumentMapping()
	commentFieldNameMapping := bleve.NewTextFieldMapping()
	updateMapping.AddFieldMappingsAt("comment", commentFieldNameMapping)
	mapping.AddDocumentMapping("update", updateMapping)
	return mapping
}

// issueIDToKey turns a numeric issue number into an internal search ID.
func issueIDToKey(id int64) string {
	return fmt.Sprintf("issue/v1/%d", id)
}

// keyToIssueID tries to convert an internal serach ID into an issue number. It
// returns 0 if the given internal ID could not be parsed as an issue ID.
func keyToIssueID(id string) int64 {
	if !strings.HasPrefix(id, "issue/v1/") {
		return 0
	}

	val, err := strconv.ParseInt(id[len("issue/v1/"):], 10, 64)
	if err != nil {
		return 0
	}
	return val
}
