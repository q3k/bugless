// Copyright 2019 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: ISC

package main

import (
	"context"
	"flag"

	"code.hackerspace.pl/hscloud/go/mirko"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/db"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	db db.Database
}

func (s *service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	return status.Error(codes.Unimplemented, "unimplemented in dummy service")
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

	db, err := db.Connect(ctx, "cockroach://q3k@185.236.240.54:26257/bugless-q3k?sslmode=require&sslrootcert=certs/cockroach-ca.crt&sslcert=certs/cockroach-client.crt&sslkey=certs/cockroach-client.key")
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
