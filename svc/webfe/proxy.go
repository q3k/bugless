// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"

	pb "github.com/q3k/bugless/proto/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type backendProxy struct {
}

func (b *backendProxy) GetIssues(req *pb.ModelGetIssuesRequest, srv pb.Model_GetIssuesServer) error {

	return status.Error(codes.Unimplemented, "unimplemented")
}

func (b *backendProxy) NewIssue(ctx context.Context, req *pb.ModelNewIssueRequest) (*pb.ModelNewIssueResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
