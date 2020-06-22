// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	cpb "github.com/q3k/bugless/proto/common"

	"github.com/inconshreveable/log15"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IssueError error

var (
	IssueErrorNotFound = status.Error(codes.NotFound, "issue not found")
)

type Issue struct {
	// Constant columns
	ID       int64  `db:"id"`
	AuthorID string `db:"author_id"`
	Created  int64  `db:"created"`

	// Bumped when a new update is added
	LastUpdated int64 `db:"last_updated"`

	// Denormalized data
	Title      string `db:"title"`
	AssigneeID string `db:"assignee_id"`
	Type       int64  `db:"type"`
	Priority   int64  `db:"priority"`
	Status     int64  `db:"status"`
}

func (i *Issue) Proto() *cpb.Issue {
	assignee := &cpb.User{Id: i.AssigneeID}
	if assignee.Id == UnassignedUUID {
		assignee = nil
	}
	return &cpb.Issue{
		Id:      i.ID,
		Created: &cpb.Timestamp{Nanos: i.Created},
		Author:  &cpb.User{Id: i.AuthorID},
		Current: &cpb.IssueState{
			Title:    i.Title,
			Assignee: assignee,
			Type:     cpb.IssueType(i.Type),
			// TODO(q3k): return CC list
			Priority: i.Priority,
			Status:   cpb.IssueStatus(i.Status),
		},
		LastUpdated: &cpb.Timestamp{Nanos: i.LastUpdated},
	}
}

// ProtoWithUsers returns a proto representation of the Issue database object
// like .Proto, but with full user data. If an error is returned, the .Proto
// result is returned (without full user data) alongside the error.
func (i *Issue) ProtoWithUsers(s Session) (*cpb.Issue, error) {
	p := i.Proto()
	author, err := s.User().Get(i.AuthorID)
	if err != nil {
		return p, err
	}
	p.Author = author.Proto()
	if p.Current.Assignee != nil {
		assignee, err := s.User().Get(p.Current.Assignee.Id)
		if err != nil {
			return p, err
		}
		p.Current.Assignee = assignee.Proto()
	}

	return p, nil
}

type IssueUpdate struct {
	IssueID  int64          `db:"issue_id"`
	UpdateID int64          `db:"id"`
	Created  int64          `db:"created"`
	AuthorID string         `db:"author_id"`
	Comment  sql.NullString `db:"comment"`

	Title      sql.NullString `db:"title"`
	AssigneeID sql.NullString `db:"assignee_id"`
	Type       sql.NullInt64  `db:"type"`
	Priority   sql.NullInt64  `db:"priority"`
	Status     sql.NullInt64  `db:"status"`
}

func (u *IssueUpdate) Proto() *cpb.Update {
	update := &cpb.Update{
		Created: &cpb.Timestamp{Nanos: u.Created},
		Author:  &cpb.User{Id: u.AuthorID},
		Comment: u.Comment.String,
		Diff:    &cpb.IssueStateDiff{},
	}

	if u.Title.Valid {
		update.Diff.Title = &cpb.IssueStateDiff_MaybeString{Value: u.Title.String}
	}
	if u.AssigneeID.Valid {
		if u.AssigneeID.String == UnassignedUUID {
			update.Diff.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: nil}
		} else {
			update.Diff.Assignee = &cpb.IssueStateDiff_MaybeUser{Value: &cpb.User{Id: u.AssigneeID.String}}
		}
	}
	if u.Type.Valid {
		update.Diff.Type = cpb.IssueType(u.Type.Int64)
	}
	if u.Priority.Valid {
		update.Diff.Priority = &cpb.IssueStateDiff_MaybeInt64{Value: u.Priority.Int64}
	}
	if u.Status.Valid {
		update.Diff.Status = cpb.IssueStatus(u.Status.Int64)
	}

	return update
}

func (u *IssueUpdate) ProtoWithUsers(s Session) (*cpb.Update, error) {
	p := u.Proto()
	author, err := s.User().Get(u.AuthorID)
	if err != nil {
		return p, err
	}
	p.Author = author.Proto()
	if u.AssigneeID.Valid && u.AssigneeID.String != UnassignedUUID {
		assignee, err := s.User().Get(u.AssigneeID.String)
		if err != nil {
			return p, err
		}
		p.Diff.Assignee.Value = assignee.Proto()
	}

	return p, nil
}

type IssueGetHistoryOpts struct {
	Start int64
	Count int64
}

type IssueFilter struct {
	// The filter passes when all the set fields match an issue.
	Author   string
	Assignee string
	Status   int64
}

type IssueOrderBy struct {
	Ascending bool
	By        IssueOrder
}

type IssueOrder int

const (
	IssueOrderCreated IssueOrder = iota
	IssueOrderUpdated
)

type IssueFilterOpts struct {
	Start int64
	Count int64
}

type IssueGetter interface {
	Get(id int64) (*Issue, error)
	Filter(filter IssueFilter, order IssueOrderBy, opts *IssueFilterOpts) ([]*Issue, error)
	GetHistory(id int64, opts *IssueGetHistoryOpts) ([]*IssueUpdate, error)
	New(new *Issue) (*Issue, error)
	Update(update *IssueUpdate) error
}

type databaseIssue struct {
	*session
}

func (d *databaseIssue) Get(id int64) (*Issue, error) {
	conv := NewErrorConverter()

	var data []*Issue

	q := `
		SELECT
			issues.id AS id,
			issues.author_id AS author_id,
			issues.created AS created,
			issues.last_updated AS last_updated,

			issues.title AS title,
			issues.assignee_id AS assignee_id,
			issues."type" AS "type",
			issues.priority AS priority,
			issues.status AS status
		FROM
			issues
		WHERE
			id = $1
	`

	err := d.tx.SelectContext(d.ctx, &data, q, id)
	if err != nil {
		return nil, conv.Convert(err)
	}

	if len(data) != 1 {
		return nil, IssueErrorNotFound
	}

	return data[0], nil
}

func (d *databaseIssue) GetHistory(id int64, opts *IssueGetHistoryOpts) ([]*IssueUpdate, error) {

	q := `
		SELECT
			issue_updates.id AS id,
			issue_updates.created AS created,
			issue_updates.author_id AS author_id,
			issue_updates.comment AS comment,
			issue_updates.title AS title,
			issue_updates.assignee_id AS assignee_id,
			issue_updates.type AS type,
			issue_updates.priority AS priority,
			issue_updates.status AS status
		FROM
			issue_updates
		WHERE
			issue_updates.issue_id = $1
	`

	if opts != nil {
		if opts.Start > 0 {
			q += fmt.Sprintf("AND issue_updates.id > %d", opts.Start)
		}
		if opts.Count > 0 {
			q += fmt.Sprintf("LIMIT %d", opts.Count)
		}
	}

	var data []*IssueUpdate
	conv := NewErrorConverter()
	err := d.tx.SelectContext(d.ctx, &data, q, id)
	if err != nil {
		return nil, conv.Convert(err)
	}

	return data, nil
}

func (d *databaseIssue) Filter(filter IssueFilter, order IssueOrderBy, opts *IssueFilterOpts) ([]*Issue, error) {
	q := `
		SELECT
			issues.id AS id,
			issues.author_id AS author_id,
			issues.created AS created,
			issues.last_updated AS last_updated,

			issues.title AS title,
			issues.assignee_id AS assignee_id,
			issues."type" AS "type",
			issues.priority AS priority,
			issues.status AS status
		FROM
			issues
	`

	var conditions []string
	var parameters []interface{}
	if filter.Author != "" {
		parameters = append(parameters, filter.Author)
		conditions = append(conditions, fmt.Sprintf("issues.author_id = $%d", len(parameters)))
	}
	if filter.Assignee != "" {
		parameters = append(parameters, filter.Assignee)
		conditions = append(conditions, fmt.Sprintf("issues.assignee_id = $%d", len(parameters)))
	}
	if filter.Status != 0 {
		parameters = append(parameters, filter.Status)
		conditions = append(conditions, fmt.Sprintf("issues.status = $%d", len(parameters)))
	}

	orderField := "issues.created"
	if opts != nil && opts.Start > 0 {
		switch order.By {
		case IssueOrderCreated:
			orderField = "issues.created"
		case IssueOrderUpdated:
			orderField = "issues.last_updated"
		default:
			return nil, status.Errorf(codes.InvalidArgument, "invalid order")
		}
		parameters = append(parameters, opts.Start)
		if order.Ascending {
			conditions = append(conditions, fmt.Sprintf("%s > $%d", orderField, len(parameters)))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s < $%d", orderField, len(parameters)))
		}
	}

	if len(conditions) > 0 {
		q += fmt.Sprintf(`
			WHERE
				%s
		`, strings.Join(conditions, " AND "))
	}

	if order.Ascending {
		q += fmt.Sprintf(`
			ORDER BY %s ASC
		`, orderField)
	} else {
		q += fmt.Sprintf(`
			ORDER BY %s DESC
		`, orderField)
	}

	if opts != nil && opts.Count > 0 {
		parameters = append(parameters, opts.Count)
		q += fmt.Sprintf(`
			LIMIT $%d
		`, len(parameters))
	}

	var data []*Issue
	conv := NewErrorConverter().
		WithSyntaxError(UserErrorNoSuchUsername)
	err := d.tx.SelectContext(d.ctx, &data, q, parameters...)
	if err != nil {
		return nil, conv.Convert(err)
	}

	return data, nil
}

func (d *databaseIssue) New(new *Issue) (*Issue, error) {
	if new.ID != 0 {
		return nil, status.Error(codes.InvalidArgument, "issue cannot contain preset id")
	}
	if new.Created == 0 {
		new.Created = time.Now().UnixNano()
	}
	new.LastUpdated = time.Now().UnixNano()
	if new.Created > new.LastUpdated {
		return nil, status.Errorf(codes.InvalidArgument, "issue creation time cannot be after last update time")
	}

	conv := NewErrorConverter()
	q := `
		INSERT INTO issues
			(author_id, created, last_updated,
			 title, assignee_id, "type", priority, status)
		VALUES
			(:author_id, :created, :last_updated,
			 :title, :assignee_id, :type, :priority, :status)
		RETURNING id
	`
	data := *new
	if data.AssigneeID == "" {
		data.AssigneeID = UnassignedUUID
	}

	rows, err := d.tx.NamedQuery(q, &data)
	if err != nil {
		return nil, conv.Convert(err)
	}
	// Get new ID
	if !rows.Next() {
		rows.Close()
		return nil, status.Error(codes.Unavailable, "could not create issue")
	}
	var id int64
	if err = rows.Scan(&id); err != nil {
		rows.Close()
		return nil, conv.Convert(err)
	}
	rows.Close()

	log15.Info("created new issue", "id", id)
	data.ID = id

	return &data, nil
}

func (d *databaseIssue) Update(update *IssueUpdate) error {
	conv := NewErrorConverter()
	now := time.Now().UnixNano()

	data := *update
	data.Created = now
	if data.AssigneeID.Valid && data.AssigneeID.String == "" {
		data.AssigneeID.String = UnassignedUUID
	}

	updates := []string{"last_updated"}
	args := []interface{}{now}

	if data.Title.Valid {
		updates = append(updates, "title")
		args = append(args, data.Title.String)
	}
	if data.AssigneeID.Valid {
		updates = append(updates, "assignee_id")
		args = append(args, data.AssigneeID)
	}
	if data.Type.Valid {
		updates = append(updates, "type")
		args = append(args, data.Type.Int64)
	}
	if data.Priority.Valid {
		updates = append(updates, "priority")
		args = append(args, data.Priority.Int64)
	}
	if data.Status.Valid {
		updates = append(updates, "status")
		args = append(args, data.Status.Int64)
	}

	var updateStrings []string
	for i, u := range updates {
		updateStrings = append(updateStrings, fmt.Sprintf("%s = $%d", u, i+1))
	}

	q := `
		UPDATE issues
		SET
			`
	q += strings.Join(updateStrings, ", ")
	q += fmt.Sprintf(" WHERE id = $%d", len(args)+1)
	fmt.Println(q)
	fmt.Println(args)
	args = append(args, data.IssueID)
	_, err := d.tx.Exec(q, args...)
	if err != nil {
		return conv.Convert(err)
	}

	q = `
		INSERT INTO issue_updates
			(issue_id, created, author_id, comment,
			 title, assignee_id, type, priority, status,
			 id)
		VALUES
			(:issue_id, :created, :author_id, :comment,
			 :title, :assignee_id, :type, :priority, :status,
			 (
			   SELECT COUNT(*)+1 from issue_updates where issue_id = :issue_id
			 )
			)
	`
	_, err = d.tx.NamedExec(q, &data)
	if err != nil {
		return conv.Convert(err)
	}

	return nil
}
