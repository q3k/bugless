// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/service"

	log "github.com/inconshreveable/log15"
)

var (
	flagEatMyData bool
	flagDSN       string
)

func main() {
	flag.BoolVar(&flagEatMyData, "eat_my_data", false, "Run crdb model again an in-memory database. This will be cleared on shutdown, use this for development purposes only")
	flag.StringVar(&flagDSN, "dsn", "", "DSN, like cockroach://user@host:port/database?sslmode=require&sslrootcert=...")
	flag.Parse()
	m := mirko.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	ctx := m.Context()

	var s *service.Service
	var err error

	if flagEatMyData {
		l.Warn("Running with in-memory database. This WILL EAT YOUR DATA")
		s, err = service.NewInMemory(ctx, l)
	} else {
		s, err = service.NewDSN(ctx, flagDSN, true, l)
	}
	if err != nil {
		l.Crit("creating service failed", "err", err)
		return
	}

	spb.RegisterModelServer(m.GRPC(), s)

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}
