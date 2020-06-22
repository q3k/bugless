// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"database/sql"
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

	// Begin returns a database Session that must be commited, but can also be rolled back
	Begin(ctx context.Context) Session
	// Do returns a database Session that will automatically commit on every object access
	Do(ctx context.Context) Session
}

type Session interface {
	Category() CategoryGetter
	Issue() IssueGetter
	User() UserGetter
	Commit() error
	Rollback() error
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

	return res, nil
}

type database struct {
	db           *sqlx.DB
	dsnPostgres  string
	dsnCockroach string
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

func (d *database) Begin(ctx context.Context) Session {
	tx := d.db.MustBeginTx(ctx, &sql.TxOptions{})
	res := &session{
		ctx: ctx,
		tx:  tx,
	}
	res.category = &databaseCategory{res}
	res.issue = &databaseIssue{res}
	res.user = &databaseUser{res}
	return res
}

func (d *database) Do(ctx context.Context) Session {
	return &autoSession{
		db:  d,
		ctx: ctx,
	}
}

type session struct {
	ctx      context.Context
	tx       *sqlx.Tx
	category *databaseCategory
	issue    *databaseIssue
	user     *databaseUser
}

func (s *session) Commit() error {
	return s.tx.Commit()
}

func (s *session) Rollback() error {
	return s.tx.Rollback()
}

func (s *session) Category() CategoryGetter {
	return s.category
}

func (s *session) Issue() IssueGetter {
	return s.issue
}

func (s *session) User() UserGetter {
	return s.user
}
