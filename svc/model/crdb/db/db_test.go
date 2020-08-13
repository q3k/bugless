// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
)

// Test users populated in the dut database.
var (
	testUsers = map[string]string{
		"q3k":   "f0b4cb6a-e36a-40dd-9d1d-a4490e5e9773",
		"implr": "73440ef3-f654-4b95-8500-253c90209f66",
	}
)

func dut(ctx context.Context, t *testing.T) (Database, func()) {
	ts, err := testserver.NewTestServer()
	if err != nil {
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

	// This is a handcrafted query by design - we don't want to exercise any
	// of the db_users.go code yet.
	inner := db.(*database)
	for username, uuid := range testUsers {
		_, err := inner.db.Exec(`
			INSERT INTO users (id, username, preferences)
			VALUES ($1, $2, '')
		`, uuid, username)
		if err != nil {
			panic(fmt.Errorf("could not create test user %q: %v", username, err))
		}
	}

	return db, ts.Stop
}

func commit(s Session, t *testing.T) {
	if err := s.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
}
