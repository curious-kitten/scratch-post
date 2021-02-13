package users

import (
	"context"
	"database/sql"
)

// UserDB encapsulates user queries
type UserDB interface {
	GetUser(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	GetPasswordForUser(ctx context.Context, username string) (string, error)
}

type userDB struct {
	db *sql.DB
}

// GetUser returns the user matching the username from the DB
func (u *userDB) GetUser(ctx context.Context, username string) (*User, error) {
	stmt := "select username, email, name from users WHERE username=$1"
	user := &User{}
	row := u.db.QueryRowContext(ctx, stmt, username)
	if row.Err() != nil {
		return nil, row.Err()
	}
	if err := row.Scan(&user.Username, &user.Email, &user.Name); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userDB) GetPasswordForUser(ctx context.Context, username string) (string, error) {
	stmt := "select password from users WHERE username=$1"
	var password string
	row := u.db.QueryRowContext(ctx, stmt, username)
	if row.Err() != nil {
		return "", row.Err()
	}
	if err := row.Scan(&password); err != nil {
		return "", err
	}
	return password, nil
}

func (u *userDB) CreateUser(ctx context.Context, user *User) error {
	stmt := "insert into users (username, name, email, password) values($1, $2, $3, $4)"

	row := u.db.QueryRowContext(ctx, stmt, user.Username, user.Name, user.Email, user.Password)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

// NewUserDB creates a wrapper around the queries used to perform user operations
func NewUserDB(db *sql.DB) UserDB {
	return &userDB{db}
}
