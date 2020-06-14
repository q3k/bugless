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
		if diff.Assignee.Value != nil {
			update.Assignee.String = diff.Assignee.Value.Id
		}
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

// applyUpdateLogic is a hairly ball of logic to ensure that issue states
// respect some invariants. These invariants are currently defined to be:
//  - an issue cannot be NEW and assigned to someone at the same time
//  - an issue cannot be non-NEW and not assigned to anyone at the same time
//  - an issue that was ACCEPTED and got reassigned without an explicit status
//    change should get changed to ASSIGNED.
//
// This logic could be moved to the database (ie., denormalized where NEW and
// ASSIGNED are the same state) - but we're keeping it in the application as it
// might be customizable in the future.
//
// Alternatively, we could try to express this logic as (programmable?) issue
// workflows/ lifecycles - but that's a large chunk of work that falls into the
// category of general programmability of bugless, which is in turn part of a
// larger discussion about the intended target and design of bugless. For now,
// let's hardcode all of this and be done.
func applyUpdateLogic(cur *cpb.IssueState, d *cpb.IssueStateDiff) {
	// Simulate application of diff to current state.
	new := *cur
	if d.Title != nil {
		new.Title = d.Title.Value
	}
	if d.Assignee != nil {
		new.Assignee = d.Assignee.Value
	}
	if validation.IssueType(d.Type) == nil {
		new.Type = d.Type
	}
	if d.Priority != nil && validation.IssuePriority(d.Priority.Value) == nil {
		new.Priority = d.Priority.Value
	}
	if validation.IssueStatus(d.Status) == nil {
		new.Status = d.Status
	}

	if new.Status == cpb.IssueStatus_NEW && new.Assignee != nil {
		// Problem: an issue cannot be NEW and have someone assigned.

		if d.Status == cpb.IssueStatus_NEW {
			// If the NEW state is caused by the diff...
			if d.Assignee != nil && cur.Assignee == nil {
				// and the diff also assigns someone, remove the assignment.
				d.Assignee = nil
			} else {
				// otherwise, force unassignment in diff (it means the issue
				// was already assigned to someone).
				d.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
			}
		} else if cur.Status == cpb.IssueStatus_NEW && d.Status == cpb.IssueStatus_ISSUE_STATUS_INVALID && d.Assignee != nil && d.Assignee.Value != nil {
			// If the new diff tries to assign someone without changing the
			// state to ASSIGNED, do that for them.
			d.Status = cpb.IssueStatus_ASSIGNED
		} else {
			// Out of nice options for problem resolution: just force
			// unassignment of user.
			d.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
		}
	} else if new.Status != cpb.IssueStatus_NEW && new.Assignee == nil {
		// Problem: a non-NEW issue cannot be unassigned.

		if cur.Status == cpb.IssueStatus_NEW && new.Status != cpb.IssueStatus_NEW {
			// If the non-NEW status is caused by the diff...
			if cur.Assignee != nil && d.Assignee != nil {
				// and the diff also caused the unassign, remove the unassign.
				d.Assignee = nil
			} else {
				// otherwise, nuke the change to non-NEW, as we don't know
				// who to assign to.
				d.Status = cpb.IssueStatus_ISSUE_STATUS_INVALID
			}
		} else if cur.Assignee != nil && new.Assignee == nil {
			// If unassignment is caused by the diff, move the issue to NEW status.
			d.Status = cpb.IssueStatus_NEW
		} else {
			// Out of nice options for problem resolution: force NEW status
			// and unassignment.
			d.Status = cpb.IssueStatus_NEW
			d.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
		}
	}
}
