package main

import (
	"context"

	spb "github.com/q3k/bugless/proto/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	return status.Error(codes.Unimplemented, "unimplemented")
}

func (s *service) NewIssue(ctx context.Context, req *spb.ModelNewIssueRequest) (*spb.ModelNewIssueResponse, error) {

	if err := validateNewIssue(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue: %w", err)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
