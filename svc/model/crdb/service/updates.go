package service

import (
	"context"

	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/common/pagination"
	"github.com/q3k/bugless/svc/model/common/validation"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) GetIssueUpdates(req *spb.ModelGetIssueUpdatesRequest, srv spb.Model_GetIssueUpdatesServer) error {
	ctx := srv.Context()

	// TODO(q3k): define the consistency guarantees for this call.
	session := s.db.Begin(ctx)
	defer session.Rollback()

	return pagination.ResampleInt64(req.Pagination, func(first bool, start pagination.V, count int64) (int, pagination.V, error) {
		opts := &db.IssueGetHistoryOpts{Start: start.(int64), Count: count}
		updates, err := session.Issue().GetHistory(req.Id, opts)
		if err != nil {
			return 0, start, err
		}

		chunk := &spb.ModelGetIssueUpdatesChunk{}
		for _, u := range updates {
			chunk.Updates = append(chunk.Updates, u.Proto())
		}

		if len(updates) > 0 {
			start = updates[len(updates)-1].UpdateID
		}
		return len(updates), start, srv.Send(chunk)
	})
}

func (s *Service) UpdateIssue(ctx context.Context, req *spb.ModelUpdateIssueRequest) (*spb.ModelUpdateIssueResponse, error) {
	if err := validation.User(req.Author); err != nil {
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
	if validation.IssueType(req.New.Type) == nil {
		update.Type.Valid = true
		update.Type.Int64 = int64(req.New.Type)
	}
	// TODO(q3k): fix not being able to set P0, this requires a schema fix
	if req.New.Priority > 0 && validation.IssuePriority(req.New.Priority) == nil {
		update.Priority.Valid = true
		update.Priority.Int64 = req.New.Priority
	}
	if validation.IssueStatus(req.New.Status) == nil {
		update.Status.Valid = true
		update.Status.Int64 = int64(req.New.Status)
	}

	err := s.db.Do(ctx).Issue().Update(update)
	if err != nil {
		return nil, err
	}
	return &spb.ModelUpdateIssueResponse{}, nil
}
