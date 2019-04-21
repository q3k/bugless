// Copyright 2019 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: ISC

package main

import (
	"flag"

	"github.com/q3k/bugless/lib/micro"
	spb "github.com/q3k/bugless/proto/svc"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
}

func (s *service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	return status.Error(codes.Unimplemented, "unimplemented in dummy service")
}

func main() {
	flag.Parse()
	m := micro.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	s := &service{}
	spb.RegisterModelServer(m.GRPC(), s)

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}
