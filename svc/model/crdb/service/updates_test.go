package service

import (
	"testing"

	cpb "github.com/q3k/bugless/proto/common"

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
			&cpb.IssueState{},
			&cpb.IssueStateDiff{},
			&cpb.IssueStateDiff{},
		},

		// Assignment to user causes status ASSIGNED
		{
			&cpb.IssueState{},
			&cpb.IssueStateDiff{
				Assignee: &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: "q3k"}},
				Status:   cpb.IssueStatus_WONTFIX_UNFORTUNATE,
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
			&cpb.IssueState{},
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
			t.Errorf("test %d: found differences.", i)
		}
	}
}
