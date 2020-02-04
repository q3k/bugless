// Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"database/sql"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/inconshreveable/log15"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CategoryError error

var (
	CategoryErrorParentNotFound   = status.Error(codes.NotFound, "parent not found")
	CategoryErrorDuplicateName    = status.Error(codes.AlreadyExists, "duplicate category name")
	CategoryErrorNotFound         = status.Error(codes.NotFound, "category not found")
	CategoryErrorCannotDeleteRoot = status.Error(codes.InvalidArgument, "cannot delete root category")
	CategoryErrorNotEmpty         = status.Error(codes.FailedPrecondition, "category has dependent data")
)

// The UUID of the root of the category tree.
const RootCategory = "00000000-0000-0000-0000-000000000000"

// An issue category.
type Category struct {
	UUID       string `db:"id"`
	ParentUUID string `db:"parent_id"`

	Name        string `db:"name"`
	Description string `db:"description"`
}

// A category that's part of a retrieved category tree.
type CategoryNode struct {
	*Category

	Children []*CategoryNode
}

type CategoryGetter interface {
	// Get retrieves a database category by UUID.
	Get(ctx context.Context, uuid string) (*Category, error)
	// GetTree returns a (sub)tree of categories starting at a given category
	// UUID.
	// At most `levels`  category tree levels. Level 0 means only the requested
	// node, level 1 the node and its children, level 2 the node, it's children
	// and it's grandchildren, etc.
	GetTree(ctx context.Context, rootUUID string, levels uint) (*CategoryNode, error)
	// New creates a new database category from an in-memroy category.  UUID
	// must be unset. ParentUUID should either point to an existing category
	// (if the category is a chuild category) or be blank (if the category
	// should be to-level).
	New(ctx context.Context, new *Category) (*Category, error)
	// Update saves a given category. All fields can be updated apart from the
	// current UUID.
	Update(ctx context.Context, cat *Category) error
	// Delete removes a category. It must not contain any child categories or
	// issues.
	Delete(ctx context.Context, uuid string) error
}

func (d *databaseCategory) Get(ctx context.Context, uuid string) (*Category, error) {
	conv := NewErrorConverter().
		WithSyntaxError(CategoryErrorNotFound)

	data := []*Category{}

	q := `
		SELECT
			categories.id AS id,
			IF(categories.parent_id IS NULL, '', categories.parent_id::string) AS parent_id,
			categories.name AS name,
			categories.description AS description
		FROM
			categories
		WHERE
			id = $1
	`

	err := d.db.SelectContext(ctx, &data, q, uuid)
	if err != nil {
		return nil, conv.Convert(err)
	}

	if len(data) != 1 {
		return nil, CategoryErrorNotFound
	}

	return data[0], nil
}

func (d *databaseCategory) GetTree(ctx context.Context, rootUUID string, levels uint) (*CategoryNode, error) {
	// We retrieve the category tree in application code.
	// CockroachDB has no suppport for recursive queries, and categories are
	// not our main load. As such, this is good enough.
	tx := d.db.MustBeginTx(ctx, &sql.TxOptions{})
	defer tx.Rollback()

	elems := make(map[string]*CategoryNode)
	levelsM := make(map[string]uint)

	// Get root.
	elem, err := d.Get(ctx, rootUUID)
	if err != nil {
		return nil, err
	}
	elems[rootUUID] = &CategoryNode{Category: elem, Children: []*CategoryNode{}}
	levelsM[rootUUID] = 0

	// Short circuit for levels == 0.
	if levels == 0 {
		return elems[rootUUID], nil
	}

	// BFS through tree.
	queue := []string{rootUUID}

	for {
		if len(queue) == 0 {
			break
		}
		uuid := queue[0]
		queue = queue[1:]

		q := `
			SELECT
				categories.id AS id,
				categories.parent_id as parent_id,
				categories.name AS name,
				categories.description AS description
			FROM
				categories
			WHERE
				parent_id = $1
		`

		data := []*Category{}
		err := d.db.SelectContext(ctx, &data, q, uuid)
		if err != nil {
			return nil, err
		}

		for _, datum := range data {
			this := &CategoryNode{Category: datum, Children: []*CategoryNode{}}
			parent := elems[uuid]
			elems[datum.UUID] = this

			parent.Children = append(parent.Children, this)

			level := levelsM[uuid] + 1
			levelsM[datum.UUID] = level
			if level >= levels {
				continue
			}

			queue = append(queue, datum.UUID)
		}
	}

	return elems[rootUUID], nil
}

func (d *databaseCategory) New(ctx context.Context, new *Category) (*Category, error) {
	if new.UUID != "" {
		return nil, status.Error(codes.InvalidArgument, "category cannot contain preset UUID")
	}
	if new.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "category must have a name")
	}
	if new.ParentUUID == "" {
		return nil, status.Error(codes.InvalidArgument, "category must have a parent")
	}

	conv := NewErrorConverter().
		WithSyntaxError(CategoryErrorParentNotFound).
		WithForeignKeyViolation(CategoryErrorParentNotFound).
		WithUniqueConstraintViolation(CategoryErrorDuplicateName)

	tx := d.db.MustBeginTx(ctx, &sql.TxOptions{})
	defer tx.Rollback()

	// Ensure that the parent exists. This is a weird quirk of CRDB -
	// constaints should be handling this instead!
	// TODO(q3k): investigate this.
	var count []int64
	q := `
		SELECT
			COUNT(*)
		FROM
			categories
		WHERE
			id = $1
	`
	err := tx.Select(&count, q, new.ParentUUID)
	log15.Info("dupa", "err", err, "count", count)
	if err != nil {
		return nil, conv.Convert(err)
	}

	if count[0] == 0 {
		return nil, CategoryErrorParentNotFound
	}

	// Create new category
	q = `
		INSERT INTO categories
			(parent_id, name, description)
		VALUES
			(:parent_id, :name, :description)
		RETURNING id
	`
	data := *new
	rows, err := tx.NamedQuery(q, &data)
	if err != nil {
		return nil, conv.Convert(err)
	}
	// Get new ID
	if !rows.Next() {
		rows.Close()
		return nil, status.Error(codes.Unavailable, "could not create category")
	}

	var uuid string
	if err = rows.Scan(&uuid); err != nil {
		rows.Close()
		return nil, conv.Convert(err)
	}
	rows.Close()

	log15.Info("created new category", "uuid", uuid, "name", data.Name)
	data.UUID = uuid

	return &data, tx.Commit()
}

func (d *databaseCategory) Update(ctx context.Context, cat *Category) error {
	if cat.UUID == "" {
		return status.Error(codes.InvalidArgument, "an updated category must already be saved")
	}
	if cat.Name == "" {
		return status.Error(codes.InvalidArgument, "category must have a name")
	}
	if cat.ParentUUID == "" {
		return status.Error(codes.InvalidArgument, "category must have a parent")
	}

	conv := NewErrorConverter().
		WithSyntaxError(CategoryErrorParentNotFound).
		WithForeignKeyViolation(CategoryErrorParentNotFound).
		WithUniqueConstraintViolation(CategoryErrorDuplicateName)

	q := `
		UPDATE categories
		SET
			parent_id = :parent_id,
			name = :name,
			description = :description
		WHERE
			id = :id
	`

	_, err := d.db.NamedExecContext(ctx, q, &cat)
	return conv.Convert(err)
}

func (d *databaseCategory) Delete(ctx context.Context, uuid string) error {
	if uuid == RootCategory {
		return CategoryErrorCannotDeleteRoot
	}

	conv := NewErrorConverter().
		WithSyntaxError(nil). // swallow invalid uuids
		WithForeignKeyViolation(CategoryErrorNotEmpty)

	q := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := d.db.ExecContext(ctx, q, uuid)
	return conv.Convert(err)
}
