package service

import (
	"context"
	"fmt"
	"io"
	"testing"

	cpb "github.com/q3k/bugless/proto/common"
	spb "github.com/q3k/bugless/proto/svc"
)

func TestIssueCreationSelectionStream(t *testing.T) {
	ctx := context.Background()

	model, users, cancel := dutModel()
	defer cancel()

	// Make a thousand (rounded up) issues.
	issueIds := []int64{}
	for i := 0; i < 1337; i++ {
		req := &spb.ModelNewIssueRequest{
			Author: users["implr"],
			InitialState: &cpb.IssueState{
				Title:    fmt.Sprintf("test issue %d", i),
				Type:     cpb.IssueType_BUG,
				Priority: 2,
				Status:   cpb.IssueStatus_NEW,
			},
		}
		res, err := model.NewIssue(ctx, req)
		if err != nil {
			t.Fatalf("NewIssue: %v", err)
		}
		issueIds = append(issueIds, res.Id)
	}

	// Ensure issue IDs are monotonics.
	prev := issueIds[0]
	for ix, i := range issueIds[1:] {
		if i <= prev {
			t.Fatalf("%dth issue filed as non-monotonic ID: previous was %d, this is %d", ix, prev, i)
		}
		prev = i
	}

	// Retrieve all issues by search and ensure they're all there.
	want := make(map[int64]bool)
	for _, i := range issueIds {
		want[i] = true
	}
	srv, err := model.GetIssues(ctx, &spb.ModelGetIssuesRequest{
		Query: &spb.ModelGetIssuesRequest_BySearch_{
			BySearch: &spb.ModelGetIssuesRequest_BySearch{
				Search: "author:implr",
			},
		},
		Pagination: &spb.PaginationSelector{
			Count: 1337,
		},
		OrderBy: spb.ModelGetIssuesRequest_ORDER_BY_CREATED,
	})
	if err != nil {
		t.Fatalf("GetIssues: %v", err)
	}
	for {
		chunk, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Recv: %v", err)
		}
		for _, issue := range chunk.Issues {
			delete(want, issue.Id)
		}
	}
	// Ensure all have been retrieved.
	for id, need := range want {
		if !need {
			continue
		}
		t.Fatalf("did not found issue %d in response", id)
	}
}

func TestIssueUpdating(t *testing.T) {
	ctx := context.Background()

	model, users, cancel := dutModel()
	defer cancel()

	// Create an issue.
	req := &spb.ModelNewIssueRequest{
		Author: users["implr"],
		InitialState: &cpb.IssueState{
			Title:    "test issue",
			Type:     cpb.IssueType_BUG,
			Priority: 2,
			Status:   cpb.IssueStatus_NEW,
		},
	}
	res, err := model.NewIssue(ctx, req)
	if err != nil {
		t.Fatalf("NewIssue: %v", err)
	}

	updateTitle := func(n int) string { return fmt.Sprintf("test issue %d", n) }
	updateComment := func(n int) string { return fmt.Sprintf("updating! %d", n) }

	// Send a thousand of updates or so. Note them down.
	for i := 0; i < 1337; i++ {
		req2 := &spb.ModelUpdateIssueRequest{
			Id:      res.Id,
			Author:  users["implr"],
			Comment: updateComment(i),
			Diff: &cpb.IssueStateDiff{
				Title: &cpb.IssueStateDiff_MaybeString{Value: updateTitle(i)},
			},
		}
		_, err := model.UpdateIssue(ctx, req2)
		if err != nil {
			t.Fatalf("UpdateIssue: %v", err)
		}
	}

	// Retrieve all updates.
	req2 := &spb.ModelGetIssueUpdatesRequest{
		Id:   res.Id,
		Mode: spb.ModelGetIssueUpdatesRequest_MODE_STATUS_AND_UPDATES,
		Pagination: &spb.PaginationSelector{
			Count: 1337,
		},
	}

	srv, err := model.GetIssueUpdates(ctx, req2)
	if err != nil {
		t.Fatalf("GetIssueUpdates: %v", err)
	}

	i := 0
	for {
		chunk, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Recv: %v", err)
		}

		for _, update := range chunk.Updates {
			if want, got := updateTitle(i), update.Diff.Title.Value; want != got {
				t.Fatalf("update %d: wanted title %q, got %q", want, got)
			}
			if want, got := updateComment(i), update.Comment; want != got {
				t.Fatalf("update %d: wanted comment %q, got %q", want, got)
			}
			i++
		}
	}
	if want, got := 1337, i; want != got {
		t.Fatalf("wanted %d updates, got %d", want, got)
	}
}
