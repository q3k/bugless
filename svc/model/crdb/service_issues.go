package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/common"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	switch inner := req.Query.(type) {
	case *spb.ModelGetIssuesRequest_ById_:
		return s.getIssueById(inner.ById, srv)
	case *spb.ModelGetIssuesRequest_BySearch_:
		return s.getIssuesBySearch(inner.BySearch, srv)
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

func (s *service) getIssuesBySearch(req *spb.ModelGetIssuesRequest_BySearch, srv spb.Model_GetIssuesServer) error {
	ctx := srv.Context()
	req.Search = strings.TrimSpace(req.Search)
	if req.Search == "" {
		return status.Error(codes.InvalidArgument, "search must be set and non-empty")
	}
	q := common.ParseSearch(req.Search)
	s.l.Debug("query by search", "query", q)

	// Try to parse ID, if given. Otherwise will be 0.
	var id int64
	if idStr := strings.TrimSpace(q.ID); idStr != "" {
		id_, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			id = id_
		}
	}

	// Simple case: if the ID is set to a valid number, that's just a get-by-id.
	if id != 0 {
		issue, err := s.db.Do(ctx).Issue().Get(id)
		if err != nil {
			return err
		}

		issueProto := issue.Proto()
		if issueProto == nil {
			return status.Error(codes.Internal, "database entry for issue could not be parsed")
		}

		return srv.Send(issueProto)
	}

	filter := db.IssueFilter{
		Author:   strings.ToLower(strings.TrimSpace(q.Author)),
		Assignee: strings.ToLower(strings.TrimSpace(q.Assignee)),
		Status:   int64(common.ParseIssueStatus(q.Status)),
	}
	if filter.Author == "" && filter.Assignee == "" && filter.Status == 0 {
		return status.Error(codes.Unimplemented, "no keyword search implemented, use query filters")
	}

	// TODO(q3k): expose this in proto
	orderBy := db.IssueOrderBy{
		By: db.IssueOrderUpdated,
	}
	opts := db.IssueFilterOpts{
		Start: 0,
		Count: 100,
	}

	issues, err := s.db.Do(ctx).Issue().Filter(filter, orderBy, &opts)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		issueProto := issue.Proto()
		if issueProto == nil {
			return status.Error(codes.Internal, "database entry for issue could not be parsed")
		}
		if err := srv.Send(issueProto); err != nil {
			return err
		}
	}
	return nil
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
