// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach-go/testserver"
)

func dut(ctx context.Context, t *testing.T) (Database, func()) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Start(); err != nil {
		t.Fatal(err)
	}

	dsn := "cockroach://" + strings.TrimPrefix(ts.PGURL().String(), "postgresql://")

	db, err := Connect(ctx, dsn)
	if err != nil {
		ts.Stop()
		t.Fatalf("Could not connect to database: %v", err)
	}

	if err = db.Migrate(); err != nil {
		ts.Stop()
		t.Fatalf("Could not migrate database: %v", err)
	}

	return db, ts.Stop
}
