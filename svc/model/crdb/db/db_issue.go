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
	ID      int64  `db:"id"`
	Author  string `db:"author"`
	Created int64  `db:"created"`

	// Bumped when a new update is added
	LastUpdated int64 `db:"last_updated"`

	// Denormalized data
	Title    string `db:"title"`
	Assignee string `db:"assignee"`
	Type     int64  `db:"type"`
	Priority int64  `db:"priority"`
	Status   int64  `db:"status"`
}

func (i *Issue) Proto() *cpb.Issue {
	return &cpb.Issue{
		Id:      i.ID,
		Created: &cpb.Timestamp{Nanos: i.Created},
		Author:  &cpb.User{Id: i.Author},
		Current: &cpb.IssueState{
			Title:    i.Title,
			Assignee: &cpb.User{Id: i.Assignee},
			Type:     cpb.IssueType(i.Type),
			// TODO(q3k): return CC list
			Priority: i.Priority,
			Status:   cpb.IssueStatus(i.Status),
		},
		LastUpdated: &cpb.Timestamp{Nanos: i.LastUpdated},
	}
}

type IssueUpdate struct {
	IssueID    int64          `db:"issue_id"`
	UpdateUUID string         `db:"id"`
	Created    int64          `db:"created"`
	Author     string         `db:"author"`
	Comment    sql.NullString `db:"comment"`

	Title    sql.NullString `db:"title"`
	Assignee sql.NullString `db:"assignee"`
	Type     sql.NullInt64  `db:"type"`
	Priority sql.NullInt64  `db:"priority"`
	Status   sql.NullInt64  `db:"status"`
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
			issues.author AS author,
			issues.created AS created,
			issues.last_updated AS last_updated,

			issues.title AS title,
			issues.assignee AS assignee,
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
			issue_updates.author AS author,
			issue_updates.comment AS comment,
			issue_updates.title AS title,
			issue_updates.assignee AS assignee,
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
			issues.author AS author,
			issues.created AS created,
			issues.last_updated AS last_updated,

			issues.title AS title,
			issues.assignee AS assignee,
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
		conditions = append(conditions, fmt.Sprintf("issues.author = $%d", len(parameters)))
	}
	if filter.Assignee != "" {
		parameters = append(parameters, filter.Assignee)
		conditions = append(conditions, fmt.Sprintf("issues.assignee = $%d", len(parameters)))
	}
	if filter.Status != 0 {
		parameters = append(parameters, filter.Status)
		conditions = append(conditions, fmt.Sprintf("issues.status = $%d", len(parameters)))
	}

	if opts != nil && opts.Start > 0 {
		field := "issues.created"
		switch order.By {
		case IssueOrderUpdated:
			field = "issues.last_updated"
		}
		parameters = append(parameters, opts.Start)
		if order.Ascending {
			conditions = append(conditions, fmt.Sprintf("%s > $%d", field, len(parameters)))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s < $%d", field, len(parameters)))
		}
	}

	if len(conditions) > 0 {
		q += fmt.Sprintf(`
			WHERE
				%s
		`, strings.Join(conditions, " AND "))
	}

	if order.Ascending {
		q += `
			ORDER BY issues.last_updated ASC
		`
	} else {
		q += `
			ORDER BY issues.last_updated DESC
		`
	}

	if opts != nil && opts.Count > 0 {
		parameters = append(parameters, opts.Count)
		q += fmt.Sprintf(`
			LIMIT $%d
		`, len(parameters))
	}

	var data []*Issue
	conv := NewErrorConverter()
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
			(author, created, last_updated,
			 title, assignee, "type", priority, status)
		VALUES
			(:author, :created, :last_updated,
			 :title, :assignee, :type, :priority, :status)
		RETURNING id
	`
	data := *new
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

	updates := []string{"last_updated"}
	args := []interface{}{now}

	if update.Title.Valid {
		updates = append(updates, "title")
		args = append(args, update.Title.String)
	}
	if update.Assignee.Valid {
		updates = append(updates, "assignee")
		args = append(args, update.Assignee.String)
	}
	if update.Type.Valid {
		updates = append(updates, "type")
		args = append(args, update.Type.Int64)
	}
	if update.Priority.Valid {
		updates = append(updates, "priority")
		args = append(args, update.Priority.Int64)
	}
	if update.Status.Valid {
		updates = append(updates, "status")
		args = append(args, update.Status.Int64)
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
	args = append(args, update.IssueID)
	_, err := d.tx.Exec(q, args...)
	if err != nil {
		return conv.Convert(err)
	}

	q = `
		INSERT INTO issue_updates
			(issue_id, created, author, comment,
			 title, assignee, type, priority, status,
			 id)
		VALUES
			(:issue_id, :created, :author, :comment,
			 :title, :assignee, :type, :priority, :status,
			 (
			   SELECT COUNT(*)+1 from issue_updates where issue_id = :issue_id
			 )
			)
	`
	data := *update
	data.Created = now
	_, err = d.tx.NamedExec(q, &data)
	if err != nil {
		return conv.Convert(err)
	}

	return nil
}
