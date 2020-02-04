// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"flag"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/db"

	log "github.com/inconshreveable/log15"
)

type service struct {
	db db.Database
}

func main() {
	flag.Parse()
	m := mirko.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	ctx := context.Background()

	// TODO(q3k): make this configurable
	db, err := db.Connect(ctx, "cockroach://bugless-dev@public.crdb-waw1:26257/bugless-dev?sslmode=require&sslrootcert=certs/cockroach-ca.crt&sslcert=certs/cockroach-client.crt&sslkey=certs/cockroach-client.key")
	if err != nil {
		l.Crit("could not connect to database", "err", err)
		return
	}

	if err := db.Migrate(); err != nil {
		l.Crit("could not migrate database", "err", err)
		return
	}

	s := &service{
		db: db,
	}
	spb.RegisterModelServer(m.GRPC(), s)

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}
