// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"fmt"
	"strings"

	"code.hackerspace.pl/hscloud/go/mirko"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/q3k/bugless/svc/model/crdb/db/migrations"
)

type Database interface {
	Migrate() error

	Category() CategoryGetter
	Issue() IssueGetter
}

var traceRegistered = false

func Connect(ctx context.Context, dsn string) (Database, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn cannot be empty")
	}
	if !strings.HasPrefix(dsn, "cockroach://") {
		return nil, fmt.Errorf("dsn must be cockroach://...")
	}

	if !traceRegistered {
		// Trace requests.
		mirko.TraceSQL(&pq.Driver{}, "pgx")
		traceRegistered = true
	}

	// We trick sqlx into thinking this is a postgres database.
	dsnPostgres := "postgres://" + strings.TrimPrefix(dsn, "cockroach://")

	db, err := sqlx.ConnectContext(ctx, "pgx", dsnPostgres)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	res := &database{
		db:           db,
		dsnPostgres:  dsnPostgres,
		dsnCockroach: dsn,
	}
	res.category = &databaseCategory{res}
	res.issue = &databaseIssue{res}

	return res, nil
}

type database struct {
	db           *sqlx.DB
	dsnPostgres  string
	dsnCockroach string

	category *databaseCategory
	issue    *databaseIssue
}

type databaseCategory struct {
	*database
}

type databaseIssue struct {
	*database
}

func (d *database) Migrate() error {
	mig, err := migrations.New(d.dsnCockroach)
	if err != nil {
		return err
	}
	err = mig.Up()
	switch err {
	case migrate.ErrNoChange:
		return nil
	default:
		return err
	}
}

func (d *database) Category() CategoryGetter {
	return d.category
}

func (d *database) Issue() IssueGetter {
	return d.issue
}
