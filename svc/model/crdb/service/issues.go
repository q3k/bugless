package service

import (
	"context"
	"database/sql"
	"time"

	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/common/validation"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) NewIssue(ctx context.Context, req *spb.ModelNewIssueRequest) (*spb.ModelNewIssueResponse, error) {
	if err := validation.NewIssue(req); err != nil {
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
