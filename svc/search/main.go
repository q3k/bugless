// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"flag"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blevesearch/bleve"
	log "github.com/inconshreveable/log15"
)

var (
	flagLocalStorage = "index.bleve"
)

type service struct {
	bl bleve.Index
}

func main() {
	flag.StringVar(&flagLocalStorage, "local_storage", flagLocalStorage, "Path to local storage of this search service")
	flag.Parse()
	m := mirko.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	mapping := bleve.NewIndexMapping()
	bl, err := bleve.New(flagLocalStorage, mapping)
	if err != nil {
		l.Crit("could not create bleve index", "err", err)
		return
	}

	s := &service{
		bl: bl,
	}
	spb.RegisterSearchServer(m.GRPC(), s)

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}

func (s *service) Search(req *spb.SearchRequest, srv spb.Search_SearchServer) error {
	return status.Error(codes.Unimplemented, "nope")
}

func (s *service) IndexIssue(ctx context.Context, req *spb.IndexIssueRequest) (*spb.IndexIssueResponse, error) {
	return nil, status.Error(codes.Unimplemented, "nope")
}
