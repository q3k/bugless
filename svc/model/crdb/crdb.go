// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"github.com/cockroachdb/cockroach-go/testserver"
	log "github.com/inconshreveable/log15"
)

type service struct {
	db db.Database
	l  log.Logger
}

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
		if flagDSN == "" {
			l.Crit("dsn must be set")
			return
		}
		if !strings.HasPrefix(flagDSN, "cockroach://") {
			l.Crit("dsn must start with cockroach://")
			return
		}

		// TODO(q3k): make this configurable
		d, err = db.Connect(ctx, flagDSN)
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
		l:  l.New("component", "service"),
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
