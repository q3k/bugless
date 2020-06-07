package main

import (
	"context"

	spb "github.com/q3k/bugless/proto/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/blevesearch/bleve"
)

type service struct {
	bl bleve.Index
}

func (s *service) IndexIssue(ctx context.Context, req *spb.IndexIssueRequest) (*spb.IndexIssueResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be valid")
	}
	if len(req.Title) == 0 {
		return &spb.IndexIssueResponse{}, nil
	}

	err := s.bl.Index(issueIDToKey(req.Id), &issue{
		Title: req.Title,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "bleve.Index: %v", err)
	}

	return &spb.IndexIssueResponse{}, nil
}

func (s *service) Query(req *spb.QueryRequest, srv spb.Search_QueryServer) error {
	if req.Query == "" {
		return nil
	}
	query := bleve.NewQueryStringQuery(req.Query)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchRequest.IncludeLocations = true
	searchResult, err := s.bl.Search(searchRequest)
	if err != nil {
		return status.Errorf(codes.Unavailable, "bleve.Query: %v", err)
	}

	res := []*spb.QueryResponse_Result{}
	for _, hit := range searchResult.Hits {
		result := &spb.QueryResponse_Result{
			LocationsByTermField: make(map[string]*spb.QueryResponse_LocationsByTerm),
		}

		for field, v := range hit.Locations {
			fieldData := ""
			if _, ok := hit.Fields[field]; ok {
				if _, ok := hit.Fields[field].(string); ok {
					fieldData = hit.Fields[field].(string)
				}
			}
			lbt := &spb.QueryResponse_LocationsByTerm{
				LocationsByTerm: make(map[string]*spb.QueryResponse_Locations),
				FieldData:       fieldData,
			}
			for term, v2 := range v {
				locs := make([]*spb.QueryResponse_Location, len(v2))
				for i, location := range v2 {
					locs[i] = &spb.QueryResponse_Location{
						Pos:   location.Pos,
						Start: location.Start,
						End:   location.End,
					}
				}
				lbt.LocationsByTerm[term] = &spb.QueryResponse_Locations{
					Locations: locs,
				}
			}
			result.LocationsByTermField[field] = lbt
		}

		issueId := keyToIssueID(hit.ID)
		if issueId != 0 {
			result.Kind = spb.QueryResponse_Result_KIND_ISSUE
			result.Payload = &spb.QueryResponse_Result_Issue{
				Issue: &spb.QueryResponse_Issue{
					Id: issueId,
				},
			}
			res = append(res, result)
		}
	}

	return srv.Send(&spb.QueryResponse{
		Results: res,
	})
}
