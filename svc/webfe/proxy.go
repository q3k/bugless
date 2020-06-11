// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"io"

	pb "github.com/q3k/bugless/proto/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type backendProxy struct {
	model pb.ModelClient
}

func (b *backendProxy) GetIssues(req *pb.ModelGetIssuesRequest, srv pb.Model_GetIssuesServer) error {

	upstream, err := b.model.GetIssues(srv.Context(), req)
	if err != nil {
		return err
	}
	for {
		issue, err := upstream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := srv.Send(issue); err != nil {
			return err
		}
	}
	return nil
}

func (b *backendProxy) GetIssueUpdates(req *pb.ModelGetIssueUpdatesRequest, srv pb.Model_GetIssueUpdatesServer) error {
	upstream, err := b.model.GetIssueUpdates(srv.Context(), req)
	if err != nil {
		return err
	}
	for {
		chunk, err := upstream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := srv.Send(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (b *backendProxy) NewIssue(ctx context.Context, req *pb.ModelNewIssueRequest) (*pb.ModelNewIssueResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
