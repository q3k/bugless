// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"os"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"

	"github.com/blevesearch/bleve"
	log "github.com/inconshreveable/log15"
)

var (
	flagLocalStorage = "index.bleve"
)

func main() {
	flag.StringVar(&flagLocalStorage, "local_storage", flagLocalStorage, "Path to local storage of this search service")
	flag.Parse()
	m := mirko.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	var bl bleve.Index
	var err error
	if _, err := os.Stat(flagLocalStorage); err == nil {
		log.Info("Opening index", "path", flagLocalStorage)
		bl, err = bleve.Open(flagLocalStorage)
	} else {
		log.Info("Creating new index", "path", flagLocalStorage)
		mapping := createMapping()
		bl, err = bleve.New(flagLocalStorage, mapping)
	}
	if err != nil {
		l.Crit("could not create or open bleve index", "err", err)
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
