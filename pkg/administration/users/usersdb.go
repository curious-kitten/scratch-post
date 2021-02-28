package users

import (
	"context"
	"database/sql"

	// embeding sql queries
	_ "embed"
)

//go:embed sql/postgress/selectUser.sql
var getUserByNameSQL string

//go:embed sql/postgress/getUserPassword.sql
var getUserPasswordSQL string

//go:embed sql/postgress/insertUser.sql
var insertUserSQL string

//go:embed sql/postgress/create_users_table.sql
var initUserTableSQL string

//go:embed sql/postgress/create_session_table.sql
var initSessionTableSQL string

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
	user := &User{}
	row := u.db.QueryRowContext(ctx, getUserByNameSQL, username)
	if row.Err() != nil {
		return nil, row.Err()
	}
	if err := row.Scan(&user.Username, &user.Email, &user.Name); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userDB) GetPasswordForUser(ctx context.Context, username string) (string, error) {
	var password string
	row := u.db.QueryRowContext(ctx, getUserPasswordSQL, username)
	if row.Err() != nil {
		return "", row.Err()
	}
	if err := row.Scan(&password); err != nil {
		return "", err
	}
	return password, nil
}

func (u *userDB) CreateUser(ctx context.Context, user *User) error {
	row := u.db.QueryRowContext(ctx, insertUserSQL, user.Username, user.Name, user.Email, user.Password)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

// NewUserDB creates a wrapper around the queries used to perform user operations
func NewUserDB(db *sql.DB) (UserDB, error) {
	_, err := db.ExecContext(context.Background(), initUserTableSQL)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(context.Background(), initSessionTableSQL)
	if err != nil {
		return nil, err
	}
	return &userDB{db}, nil
}
