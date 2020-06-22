package db

import (
	"database/sql"

	cpb "github.com/q3k/bugless/proto/common"

	"github.com/inconshreveable/log15"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserError error

var (
	UserErrorNoSuchUsername    = status.Error(codes.NotFound, "no such username")
	UserErrorNoSuchUser        = status.Error(codes.NotFound, "no such user")
	UserErrorDuplicateUsername = status.Error(codes.AlreadyExists, "duplicate username")
)

// The UUID of the root of the 'unassigned' user.
const (
	UnassignedUUID = "00000000-0000-0000-0000-000000000000"
)

type User struct {
	ID          string `db:"id"`
	Username    string `db:"username"`
	Preferences []byte `db:"preferences"`

	Email       sql.NullString `db:"email"`
	DisplayName sql.NullString `db:"display_name"`
}

func (u *User) Proto() *cpb.User {
	return &cpb.User{
		Id:       u.ID,
		Username: u.Username,
	}
}

type UserGetter interface {
	New(new *User) (*User, error)
	// ResolveUsername resolves a username to its UUID.
	ResolveUsername(username string) (uuid string, err error)
	Get(uuid string) (*User, error)
}

type databaseUser struct {
	*session
}

func (d *databaseUser) New(new *User) (*User, error) {
	if new.ID != "" {
		return nil, status.Error(codes.InvalidArgument, "user cannot contain preset id")
	}
	conv := NewErrorConverter().
		WithUniqueConstraintViolation(UserErrorDuplicateUsername)

	q := `
		INSERT INTO users
			(username, preferences)
		VALUES
			(:username, '')
		RETURNING id
	`
	data := *new
	// TOOD(q3k): move to NamedQueryContext when available
	rows, err := d.tx.NamedQuery(q, &data)
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

	log15.Info("created new user", "uuid", uuid, "username", data.Username)
	data.ID = uuid

	return &data, nil
}

func (d *databaseUser) ResolveUsername(username string) (string, error) {
	q := `
		SELECT
			users.id AS id
		FROM
			users
		WHERE
			users.username = $1
	`
	var data []string
	conv := NewErrorConverter()
	err := d.tx.SelectContext(d.ctx, &data, q, username)
	if err != nil {
		return "", conv.Convert(err)
	}
	if len(data) != 1 {
		return "", UserErrorNoSuchUsername
	}
	return data[0], nil
}

func (d *databaseUser) Get(uuid string) (*User, error) {
	conv := NewErrorConverter().
		WithSyntaxError(UserErrorNoSuchUser)

	var data []*User
	q := `
		SELECT
			users.id AS id,
			users.username AS username,
			users.preferences AS preferences,
			users.email AS email,
			users.display_name as display_name
		FROM
			users
		WHERE
			id = $1
	`
	err := d.tx.SelectContext(d.ctx, &data, q, uuid)
	if err != nil {
		return nil, conv.Convert(err)
	}

	if len(data) != 1 {
		return nil, UserErrorNoSuchUser
	}

	return data[0], nil
}
