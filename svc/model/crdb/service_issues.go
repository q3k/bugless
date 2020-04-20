package main

import (
	"context"
	"database/sql"
	"time"

	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	switch inner := req.Query.(type) {
	case *spb.ModelGetIssuesRequest_ById_:
		return s.getIssueById(inner.ById, srv)
	default:
		return status.Errorf(codes.Unimplemented, "unimplemented query type %v", req.Query)
	}
}

func (s *service) getIssueById(req *spb.ModelGetIssuesRequest_ById, srv spb.Model_GetIssuesServer) error {
	ctx := srv.Context()
	issue, err := s.db.Do(ctx).Issue().Get(req.Id)
	if err != nil {
		return err
	}

	issueProto := issue.Proto()
	if issueProto == nil {
		return status.Error(codes.Internal, "database entry for issue could not be parsed")
	}

	return srv.Send(issueProto)
}

func (s *service) NewIssue(ctx context.Context, req *spb.ModelNewIssueRequest) (*spb.ModelNewIssueResponse, error) {

	if err := validateNewIssue(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid issue: %v", err)
	}

	session := s.db.Begin(ctx)
	defer session.Rollback()

	i := req.InitialState

	now := time.Now()

	assignee := ""
	if i.Assignee != nil {
		assignee = i.Assignee.Id
	}
	issue, err := session.Issue().New(&db.Issue{
		Reporter:    req.Creator.Id,
		Created:     now.UnixNano(),
		LastUpdated: now.UnixNano(),
		Title:       i.Title,
		Assignee:    assignee,
		Type:        int64(i.Type),
		Priority:    i.Priority,
		Status:      int64(i.Status),
	})

	if err != nil {
		return nil, err
	}

	if req.InitialComment != "" {
		err = session.Issue().Update(&db.IssueUpdate{
			IssueID: issue.ID,
			Comment: sql.NullString{req.InitialComment, true},
		})
		if err != nil {
			return nil, err
		}
	}

	res := &spb.ModelNewIssueResponse{
		Id: issue.ID,
	}

	return res, session.Commit()
}
