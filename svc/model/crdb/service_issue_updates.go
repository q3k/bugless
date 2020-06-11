package main

import (
	spb "github.com/q3k/bugless/proto/svc"
	"github.com/q3k/bugless/svc/model/common/pagination"
	"github.com/q3k/bugless/svc/model/crdb/db"
)

func (s *service) GetIssueUpdates(req *spb.ModelGetIssueUpdatesRequest, srv spb.Model_GetIssueUpdatesServer) error {
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
