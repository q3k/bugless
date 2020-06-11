package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/q3k/bugless/svc/model/crdb/db"

	"github.com/cockroachdb/cockroach-go/testserver"
	log "github.com/inconshreveable/log15"
)

type Service struct {
	db db.Database
	l  log.Logger
}

func NewDSN(ctx context.Context, dsn string, migrate bool, l log.Logger) (*Service, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn must be set")
	}
	if !strings.HasPrefix(dsn, "cockroach://") {
		return nil, fmt.Errorf("dsn must start with cockroach://")
	}

	d, err := db.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect: %v", err)
	}

	if migrate {
		if err := d.Migrate(); err != nil {
			return nil, fmt.Errorf("could not migrate database: %v", err)
		}
	}

	return &Service{
		db: d,
		l:  l.New("component", "service"),
	}, nil
}

func NewInMemory(ctx context.Context, l log.Logger) (*Service, error) {
	d, err := inMemoryDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create in-memory database: %v", err)
	}
	if err := d.Migrate(); err != nil {
		return nil, fmt.Errorf("could not migrate database: %v", err)
	}
	return &Service{
		db: d,
		l:  l.New("component", "service"),
	}, nil
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
