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
		return s.getIssuesBySearch(req, inner.BySearch, srv)
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

func (s *service) getIssuesBySearch(req *spb.ModelGetIssuesRequest, search *spb.ModelGetIssuesRequest_BySearch, srv spb.Model_GetIssuesServer) error {
	ctx := srv.Context()

	// Normalize request count, if zero. This thas no upper bound, as it's a streaming API.
	if req.Count <= 0 {
		req.Count = 100
	}

	search.Search = strings.TrimSpace(search.Search)
	if search.Search == "" {
		return status.Error(codes.InvalidArgument, "search must be set and non-empty")
	}
	q := common.ParseSearch(search.Search)
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
		return s.getIssueById(&spb.ModelGetIssuesRequest_ById{Id: id}, srv)
	}

	filter := db.IssueFilter{
		Author:   strings.ToLower(strings.TrimSpace(q.Author)),
		Assignee: strings.ToLower(strings.TrimSpace(q.Assignee)),
		Status:   int64(common.ParseIssueStatus(q.Status)),
	}
	if filter.Author == "" && filter.Assignee == "" && filter.Status == 0 {
		return status.Error(codes.Unimplemented, "no keyword search implemented, use query filters")
	}

	orderBy := db.IssueOrderBy{
		Ascending: true,
	}
	switch req.OrderBy {
	case spb.ModelGetIssuesRequest_ORDER_BY_CREATED:
		orderBy.By = db.IssueOrderCreated
	case spb.ModelGetIssuesRequest_ORDER_BY_LAST_UPDATE:
		orderBy.By = db.IssueOrderUpdated
	default:
		return status.Errorf(codes.InvalidArgument, "invalid order_by")
	}

	maxChunkSize := int64(100)
	var start int64
	if req.After != nil && req.After.Nanos > 0 {
		start = req.After.Nanos
	}
	var sent int64
	for {
		if sent >= req.Count {
			break
		}

		chunkSize := req.Count - sent
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		opts := db.IssueFilterOpts{
			Start: start,
			Count: chunkSize,
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

		// No more issues to send.
		if int64(len(issues)) < chunkSize {
			return nil
		}
		start = issues[len(issues)-1].Created
		sent += int64(len(issues))
	}

	return nil
}

func (s *service) GetIssueUpdates(req *spb.ModelGetIssueUpdatesRequest, srv spb.Model_GetIssueUpdatesServer) error {
	ctx := srv.Context()

	// Normalize request count, if zero. This thas no upper bound, as it's a streaming API.
	if req.Count <= 0 {
		req.Count = 100
	}

	// TODO(q3k): define the consistency guarantees for this call.
	session := s.db.Begin(ctx)
	defer session.Rollback()

	maxChunkSize := int64(100)

	var sent int64
	var start int64
	if req.After != nil && req.After.Nanos >= 0 {
		start = req.After.Nanos
	}

	first := false
	for {
		if sent >= req.Count {
			break
		}

		chunkSize := req.Count - sent
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		opts := &db.IssueGetHistoryOpts{Start: start, Count: chunkSize}
		updates, err := session.Issue().GetHistory(req.Id, opts)
		if err != nil {
			return err
		}

		chunk := &spb.ModelGetIssueUpdatesChunk{}

		if first {
			first = false
			//chunk.Current = issue
		}

		for _, u := range updates {
			chunk.Updates = append(chunk.Updates, u.Proto())
		}

		if err = srv.Send(chunk); err != nil {
			return err
		}

		if int64(len(updates)) < chunkSize {
			// No more data after this chunk.
			return nil
		}

		sent += int64(len(updates))
		start = updates[len(updates)-1].UpdateID
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
