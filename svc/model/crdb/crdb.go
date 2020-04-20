// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"code.hackerspace.pl/hscloud/go/mirko"
	"github.com/cockroachdb/cockroach-go/testserver"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/db"

	log "github.com/inconshreveable/log15"
)

type service struct {
	db db.Database
}

var (
	flagEatMyData bool
)

func main() {
	flag.BoolVar(&flagEatMyData, "eat_my_data", false, "Run crdb model again an in-memory database. This will be cleared on shutdown, use this for development purposes only")
	flag.Parse()
	m := mirko.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	ctx := m.Context()

	var d db.Database
	var err error
	if flagEatMyData {
		l.Warn("Running with in-memory database. This WILL EAT YOUR DATA")
		d, err = inMemoryDatabase(ctx)
		if err != nil {
			l.Crit("could not create in memory database", "err", err)
			return
		}
	} else {
		// TODO(q3k): make this configurable
		d, err = db.Connect(ctx, "cockroach://bugless-dev@public.crdb-waw1:26257/bugless-dev?sslmode=require&sslrootcert=certs/cockroach-ca.crt&sslcert=certs/cockroach-client.crt&sslkey=certs/cockroach-client.key")
		if err != nil {
			l.Crit("could not connect to database", "err", err)
			return
		}
	}

	if err := d.Migrate(); err != nil {
		l.Crit("could not migrate database", "err", err)
		return
	}

	s := &service{
		db: d,
	}
	spb.RegisterModelServer(m.GRPC(), s)

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}

func inMemoryDatabase(ctx context.Context) (db.Database, error) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		return nil, fmt.Errorf("NewTestServer: %v", err)
	}
	if err := ts.Start(); err != nil {
		return nil, fmt.Errorf("testserver.Start: %v", err)
	}

	dsn := "cockroach://" + strings.TrimPrefix(ts.PGURL().String(), "postgresql://")

	d, err := db.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to new database: %v", err)
	}

	go func() {
		<-ctx.Done()
		ts.Stop()
	}()

	return d, nil
}
