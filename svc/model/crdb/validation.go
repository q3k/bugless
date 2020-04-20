package main

import (
	"fmt"

	cpb "github.com/q3k/bugless/proto/common"
	spb "github.com/q3k/bugless/proto/svc"
)

func validateNewIssue(req *spb.ModelNewIssueRequest) error {
	if err := validateUser(req.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}
	if req.InitialState == nil {
		return fmt.Errorf("initial issue state must be set")
	}
	s := req.InitialState
	if s.Title == "" {
		return fmt.Errorf("issue title must be set")
	}
	if s.Assignee != nil {
		if err := validateUser(s.Assignee); err != nil {
			return fmt.Errorf("assignee: %w", err)
		}
	}
	for i, u := range s.Cc {
		if err := validateUser(u); err != nil {
			return fmt.Errorf("assignee[%d]: %w", i, err)
		}
	}
	if err := validateIssueType(s.Type); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if err := validateIssuePriority(s.Priority); err != nil {
		return fmt.Errorf("priority: %w", err)
	}
	if err := validateIssueStatus(s.Status); err != nil {
		return fmt.Errorf("status: %w", err)
	}
	return nil
}

func validateUser(u *cpb.User) error {
	if u == nil {
		return fmt.Errorf("must be set")
	}
	if len(u.Id) > 64 {
		return fmt.Errorf("must be shorter than 64 characters")
	}
	if len(u.Id) < 1 {
		return fmt.Errorf("must be at least one character")
	}
	// TODO: validate with authn
	return nil
}

func validateIssueType(t cpb.IssueType) error {
	for _, v := range []cpb.IssueType{
		cpb.IssueType_BUG,
		cpb.IssueType_FEATURE_REQUEST,
		cpb.IssueType_CUSTOMER_ISSUE,
		cpb.IssueType_INTERNAL_CLEANUP,
		cpb.IssueType_PROCESS,
		cpb.IssueType_VULNERABILITY,
	} {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported value %d", t)
}

func validateIssuePriority(p int64) error {
	if p >= 0 && p <= 4 {
		return nil
	}
	return fmt.Errorf("must be between P0 and P4")
}

func validateIssueStatus(s cpb.IssueStatus) error {
	for _, v := range []cpb.IssueStatus{
		cpb.IssueStatus_NEW,
		cpb.IssueStatus_ASSIGNED,
		cpb.IssueStatus_ACCEPTED,
		cpb.IssueStatus_FIXED,
		cpb.IssueStatus_FIXED_VERIFIED,
		cpb.IssueStatus_WONTFIX_NOT_REPRODUCIBLE,
		cpb.IssueStatus_WONTFIX_INTENDED,
		cpb.IssueStatus_WONTFIX_OBSOLETE,
		cpb.IssueStatus_WONTFIX_INFEASIBLE,
		cpb.IssueStatus_WONTFIX_UNFORTUNATE,
		cpb.IssueStatus_DUPLICATE,
	} {
		if s == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported value %d", s)
}
