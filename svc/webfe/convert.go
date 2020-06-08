package main

import (
	cpb "github.com/q3k/bugless/proto/common"
)

func issueStatusPretty(i cpb.IssueStatus) string {
	switch i {
	case cpb.IssueStatus_NEW:
		return "New"
	case cpb.IssueStatus_ASSIGNED:
		return "Assigned"
	case cpb.IssueStatus_ACCEPTED:
		return "Accepted"
	case cpb.IssueStatus_FIXED:
		return "Fixed"
	case cpb.IssueStatus_FIXED_VERIFIED:
		return "Fixed (Verified)"
	case cpb.IssueStatus_WONTFIX_NOT_REPRODUCIBLE:
		return "Won't fix (Not Reproducible)"
	case cpb.IssueStatus_WONTFIX_INTENDED:
		return "Won't fix (Intended Behaviour)"
	case cpb.IssueStatus_WONTFIX_OBSOLETE:
		return "Won't fix (Obsolete)"
	case cpb.IssueStatus_WONTFIX_INFEASIBLE:
		return "Won't fix (Infeasible)"
	case cpb.IssueStatus_WONTFIX_UNFORTUNATE:
		return "Won't fix (Unfortunate)"
	case cpb.IssueStatus_DUPLICATE:
		return "Duplicate"
	}
	return "Unknown"
}

func issueTypePretty(t cpb.IssueType) string {
	switch t {
	case cpb.IssueType_BUG:
		return "Bug"
	case cpb.IssueType_FEATURE_REQUEST:
		return "Feature Request"
	case cpb.IssueType_CUSTOMER_ISSUE:
		return "Customer Issue"
	case cpb.IssueType_INTERNAL_CLEANUP:
		return "Internal Cleanup"
	case cpb.IssueType_PROCESS:
		return "Process"
	case cpb.IssueType_VULNERABILITY:
		return "Vulnerability"
	}
	return ""
}
