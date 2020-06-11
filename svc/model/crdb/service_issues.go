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
		Author:      req.Author.Id,
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

func (s *service) UpdateIssue(ctx context.Context, req *spb.ModelUpdateIssueRequest) (*spb.ModelUpdateIssueResponse, error) {
	if err := validateUser(req.Author); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "author: %v", err)
	}
	if req.New == nil {
		return nil, status.Error(codes.InvalidArgument, "new state must be set")
	}

	update := &db.IssueUpdate{
		IssueID: req.Id,
		Author:  req.Author.Id,
	}
	if req.Comment != "" {
		update.Comment.Valid = true
		update.Comment.String = req.Comment
	}
	if req.New.Title != "" {
		update.Title.Valid = true
		update.Title.String = req.New.Title
	}
	if req.New.Assignee != nil {
		update.Assignee.Valid = true
		update.Assignee.String = req.New.Assignee.Id
	}
	if validateIssueType(req.New.Type) == nil {
		update.Type.Valid = true
		update.Type.Int64 = int64(req.New.Type)
	}
	// TODO(q3k): fix not being able to set P0, this requires a schema fix
	if req.New.Priority > 0 && validateIssuePriority(req.New.Priority) == nil {
		update.Priority.Valid = true
		update.Priority.Int64 = req.New.Priority
	}
	if validateIssueStatus(req.New.Status) == nil {
		update.Status.Valid = true
		update.Status.Int64 = int64(req.New.Status)
	}

	err := s.db.Do(ctx).Issue().Update(update)
	if err != nil {
		return nil, err
	}
	return &spb.ModelUpdateIssueResponse{}, nil
}
