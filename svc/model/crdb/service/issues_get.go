package service

import (
	"strconv"
	"strings"

	cpb "github.com/q3k/bugless/proto/common"
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/common"
	"github.com/q3k/bugless/svc/model/common/pagination"
	"github.com/q3k/bugless/svc/model/crdb/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) GetIssues(req *spb.ModelGetIssuesRequest, srv spb.Model_GetIssuesServer) error {
	switch inner := req.Query.(type) {
	case *spb.ModelGetIssuesRequest_ById_:
		return s.getIssueById(inner.ById, srv)
	case *spb.ModelGetIssuesRequest_BySearch_:
		return s.getIssuesBySearch(req, inner.BySearch, srv)
	default:
		return status.Errorf(codes.Unimplemented, "unimplemented query type %v", req.Query)
	}
}

func (s *Service) getIssueById(req *spb.ModelGetIssuesRequest_ById, srv spb.Model_GetIssuesServer) error {
	ctx := srv.Context()
	issue, err := s.db.Do(ctx).Issue().Get(req.Id)
	if err != nil {
		return err
	}

	issueProto := issue.Proto()
	if issueProto == nil {
		return status.Error(codes.Internal, "database entry for issue could not be parsed")
	}

	return srv.Send(&spb.ModelGetIssuesChunk{
		Issues: []*cpb.Issue{issueProto},
	})
}

func (s *Service) getIssuesBySearch(req *spb.ModelGetIssuesRequest, search *spb.ModelGetIssuesRequest_BySearch, srv spb.Model_GetIssuesServer) error {
	ctx := srv.Context()

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

	return pagination.ResampleInt64(req.Pagination, func(first bool, start pagination.V, count int64) (int, pagination.V, error) {
		opts := db.IssueFilterOpts{Start: start.(int64), Count: count}
		issues, err := s.db.Do(ctx).Issue().Filter(filter, orderBy, &opts)
		if err != nil {
			return 0, start, err
		}

		chunk := &spb.ModelGetIssuesChunk{}
		for _, issue := range issues {
			ip := issue.Proto()
			if ip == nil {
				return 0, start, status.Error(codes.Internal, "database entry for issue could not be parsed")
			}
			chunk.Issues = append(chunk.Issues, ip)
		}

		if len(issues) > 0 {
			start = issues[len(issues)-1].Created
		}
		return len(issues), start, srv.Send(chunk)
	})
}
