package service

import (
	"context"
	"testing"

	cpb "github.com/q3k/bugless/proto/common"
	spb "github.com/q3k/bugless/proto/svc"

	"github.com/golang/protobuf/proto"
)

func TestUpdateLogic(t *testing.T) {
	for i, te := range []struct {
		cur   *cpb.IssueState
		orig  *cpb.IssueStateDiff
		fixed *cpb.IssueStateDiff
	}{
		// Empty to empty.
		{
			&cpb.IssueState{
				Assignee: &cpb.User{Id: "q3k"},
				Status:   cpb.IssueStatus_WONTFIX_UNFORTUNATE,
			},
			&cpb.IssueStateDiff{},
			&cpb.IssueStateDiff{},
		},

		// Assignment of NEW to user causes status ASSIGNED
		{
			&cpb.IssueState{
				Status: cpb.IssueStatus_NEW,
			},
			&cpb.IssueStateDiff{
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}},
			},
			&cpb.IssueStateDiff{
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}},
				Status:   cpb.IssueStatus_ASSIGNED,
			},
		},

		// Removal of assignee causes status NEW.
		{
			&cpb.IssueState{},
			&cpb.IssueStateDiff{
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: nil},
			},
			&cpb.IssueStateDiff{
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: nil},
				Status:   cpb.IssueStatus_NEW,
			},
		},

		// State change to NEW unassigns user.
		{
			&cpb.IssueState{
				Assignee: &cpb.User{Id: "q3k"},
			},
			&cpb.IssueStateDiff{
				Status: cpb.IssueStatus_NEW,
			},
			&cpb.IssueStateDiff{
				Status:   cpb.IssueStatus_NEW,
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: nil},
			},
		},

		// Anything other than NEW should not be valid if a usser is not assigned or being assigned.
		{
			&cpb.IssueState{
				Status:   cpb.IssueStatus_ASSIGNED,
				Assignee: &cpb.User{Id: "q3k"},
			},
			&cpb.IssueStateDiff{
				Status: cpb.IssueStatus_ACCEPTED,
			},
			&cpb.IssueStateDiff{
				Status: cpb.IssueStatus_ACCEPTED,
			},
		},
		{
			&cpb.IssueState{},
			&cpb.IssueStateDiff{
				Status:   cpb.IssueStatus_ASSIGNED,
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}},
			},
			&cpb.IssueStateDiff{
				Status:   cpb.IssueStatus_ASSIGNED,
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}},
			},
		},
	} {
		applyUpdateLogic(te.cur, te.orig)
		if !proto.Equal(te.fixed, te.orig) {
			t.Errorf("test %d:  got: %+v", i, te.orig)
			t.Errorf("test %d: want: %+v", i, te.fixed)
			t.Fatalf("test %d: found differences.", i)
		}
	}
}

func TestUpdateCompaction(t *testing.T) {
	ctx := context.Background()

	model, cancel := dutModel()
	defer cancel()

	for i, te := range []struct {
		start   *cpb.IssueState
		updates []*cpb.IssueStateDiff
		end     *cpb.IssueState
	}{
		// An issue with no updates should yield the same beginning state.
		{
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
			},
			[]*cpb.IssueStateDiff{},
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
			},
		},

		// Assignement to a user should change the state to ASSIGNED.
		{
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
			},
			[]*cpb.IssueStateDiff{
				{Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}}},
			},
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_ASSIGNED,
				Assignee: &cpb.User{Id: "q3k"},
			},
		},

		// Resignation from an issue should change the state to NEW.
		{
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_ASSIGNED,
				Assignee: &cpb.User{Id: "q3k"},
			},
			[]*cpb.IssueStateDiff{
				{Assignee: &cpb.IssueStateDiff_MaybeUser{Value: nil}},
			},
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
				Assignee: nil,
			},
		},

		// A typical issue lifetime story should have a happy end.
		{
			&cpb.IssueState{
				Title:    "foo",
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
			},
			[]*cpb.IssueStateDiff{
				// Issue is filed.
				{
					Title:    &cpb.IssueStateDiff_MaybeString{Value: "foo in bar"},
					Priority: &cpb.IssueStateDiff_MaybeInt64{Value: 1},
				},

				// q3k fixes it.
				{Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}}},
				{Status: cpb.IssueStatus_ACCEPTED},
				{Status: cpb.IssueStatus_FIXED},

				// someone discovers bad fix, makes it NEW again
				{Status: cpb.IssueStatus_NEW},

				// q3k re-fixes it
				{Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}}},
				{Status: cpb.IssueStatus_ACCEPTED},
				{
					Status:   cpb.IssueStatus_FIXED,
					Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "implr"}},
				},
				// implr verifies it
				{Status: cpb.IssueStatus_FIXED_VERIFIED},
			},
			&cpb.IssueState{
				Title:    "foo in bar",
				Type:     cpb.IssueType_BUG,
				Priority: 1,
				Status:   cpb.IssueStatus_FIXED_VERIFIED,
				Assignee: &cpb.User{Id: "implr"},
			},
		},
	} {
		issueReq := &spb.ModelNewIssueRequest{
			Author:       &cpb.User{Id: "test"},
			InitialState: te.start,
		}
		issue, err := model.NewIssue(ctx, issueReq)
		if err != nil {
			t.Fatalf("test %d: NewIssue: %v", i, err)
		}

		for j, update := range te.updates {
			updateReq := &spb.ModelUpdateIssueRequest{
				Id:     issue.Id,
				Author: &cpb.User{Id: "test"},
				Diff:   update,
			}
			_, err := model.UpdateIssue(ctx, updateReq)
			if err != nil {
				t.Fatalf("test %d, update %d: UpdateIssue: %v", i, j, err)
			}
		}

		end, err := model.GetIssues(ctx, &spb.ModelGetIssuesRequest{
			Query: &spb.ModelGetIssuesRequest_ById_{
				ById: &spb.ModelGetIssuesRequest_ById{
					Id: issue.Id,
				},
			},
		})
		if err != nil {
			t.Fatalf("test %d: GetIssues: %v", i, err)
		}
		endIssues, err := end.Recv()
		if err != nil {
			t.Fatalf("test %d: Recv: %v", i, err)
		}
		if len(endIssues.Issues) != 1 {
			t.Fatalf("test %d: received more than one issue", i)
		}
		endIssue := endIssues.Issues[0]

		if !proto.Equal(endIssue.Current, te.end) {
			t.Errorf("test %d:  got: %+v", i, endIssue.Current)
			t.Errorf("test %d: want: %+v", i, te.end)
			t.Fatalf("test %d: found differences.", i)
		}
	}
}
