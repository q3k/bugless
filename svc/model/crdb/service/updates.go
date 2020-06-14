package service

import (
	"context"

	cpb "github.com/q3k/bugless/proto/common"
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

	session := s.db.Begin(ctx)
	defer session.Rollback()

	// This is somewhat ugly - but in order to check some of the update logic,
	// we need to actually retrieve the current state of the issue.
	issue, err := session.Issue().Get(req.Id)
	if err != nil {
		return nil, err
	}

	applyUpdateLogic(issue.Proto().Current, diff)

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

	err = session.Issue().Update(update)
	if err != nil {
		return nil, err
	}
	return &spb.ModelUpdateIssueResponse{}, session.Commit()
}

// applyUpdateLogic applies checks to IssueStatus logic. Issue statuses are
// tied to other fields: a NEW issue must be unassigned, every other one must
// be assigned, and changing/setting the assignee must also change the status
// to ASSIGNED.
// This logic could be moved to the database (ie., denormalized where NEW and
// ASSIGNED are the same state) - but we're keeping it in the application as it
// might be customizable in the future.
func applyUpdateLogic(cur *cpb.IssueState, d *cpb.IssueStateDiff) {
	if d.Assignee != nil {
		// If someone is being assigned, this might imply a state change.
		user := d.Assignee.Value
		if user == nil || user.Id == "" {
			// Normalize empty users to unset users.
			d.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
			// The issue got deassigned - ensure the status is NEW.
			d.Status = cpb.IssueStatus_NEW
		} else {
			// Someone got assigned - ensure the new status is ASSIGNED.
			d.Status = cpb.IssueStatus_ASSIGNED
		}
	} else if validation.IssueStatus(d.Status) == nil {
		// Make sure {NEW,!NEW} issues are unassigned and assigned respectively.
		if d.Status == cpb.IssueStatus_NEW {
			// New issues cannot be assigned to anyone.
			d.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
		} else if cur.Assignee == nil && d.Assignee == nil {
			// All other states require a set assignee. If that's not the case,
			// drop the update.
			d.Status = cpb.IssueStatus_ISSUE_STATUS_INVALID
		}
	}
}
