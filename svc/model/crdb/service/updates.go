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
	if req.Diff == nil {
		return nil, status.Error(codes.InvalidArgument, "diff must be set")
	}
	diff := req.Diff

	update := &db.IssueUpdate{
		IssueID: req.Id,
		Author:  req.Author.Id,
	}
	if req.Comment != "" {
		update.Comment.Valid = true
		update.Comment.String = req.Comment
	}
	if diff.Title != nil {
		update.Title.Valid = true
		update.Title.String = diff.Title.Value
	}
	if diff.Assignee != nil {
		update.Assignee.Valid = true
		update.Assignee.String = diff.Assignee.Value.Id
	}
	if validation.IssueType(diff.Type) == nil {
		update.Type.Valid = true
		update.Type.Int64 = int64(diff.Type)
	}
	if diff.Priority != nil && validation.IssuePriority(diff.Priority.Value) == nil {
		update.Priority.Valid = true
		update.Priority.Int64 = diff.Priority.Value
	}
	if validation.IssueStatus(diff.Status) == nil {
		update.Status.Valid = true
		update.Status.Int64 = int64(diff.Status)
	}

	err := s.db.Do(ctx).Issue().Update(update)
	if err != nil {
		return nil, err
	}
	return &spb.ModelUpdateIssueResponse{}, nil
}
