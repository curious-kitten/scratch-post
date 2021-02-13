package users

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

var (
	validEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Email is used to match a valid email address
type Email string

// IsValid is the validation function for the email
func (e Email) IsValid() bool {
	return validEmail.MatchString(string(e))
}

// Password is used to match a valid password
type Password string

// IsValid is the validation function for the password
// The password constraints are:
// Must be at least 7 long
// Must include a number
// Must include a lower case letter
// Must inslude an upper case letter
// Must contain a symbol
// Cannot contain spaces
func (p Password) IsValid() bool {
	var (
		upp bool
		low bool
		num bool
		sym bool
	)
	for _, char := range p {
		switch {
		case unicode.IsUpper(char):
			upp = true
		case unicode.IsLower(char):
			low = true
		case unicode.IsNumber(char):
			num = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
		default:
			return false
		}
	}
	if upp && low && num && sym && len(p) > 6 {
		return true
	}
	return false
}

// Username is used to match a valid username
type Username string

// IsValid is used to check is a username is valid
func (u Username) IsValid() bool {
	return !(strings.ContainsAny(string(u), " \t\n\r,.<>/?;':\"\\|[]{}-=+~`!@#$%^&*()") || len(u) < 3)
}

// User represents a used of the application
type User struct {
	Username Username `json:"username"`
	Name     string   `json:"name"`
	Email    Email    `json:"email"`
	Password Password `json:"password,omitempty"`
}

// Validate checks that all user constraints are met
func (u *User) Validate() error {
	if !u.Username.IsValid() {
		return metadata.NewValidationError("provided username is not valid")
	}
	if !u.Email.IsValid() {
		return metadata.NewValidationError("provided email is not valid")
	}
	if !u.Password.IsValid() {
		return metadata.NewValidationError("password does not meet minimum requirements")
	}
	// Not much validation here. Thanks Musk :|
	if u.Name == "" {
		return metadata.NewValidationError("name is a mandatory parameter")
	}
	return nil
}

// Create is used to add a new app user
func Create(u UserDB) func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
	return func(ctx context.Context, author string, data io.Reader) (interface{}, error) {
		user := &User{}
		if err := decoder.Decode(user, data); err != nil {
			return nil, err
		}
		if err := hashPasswordForUser(user); err != nil {
			return nil, err
		}
		if err := u.CreateUser(ctx, user); err != nil {
			return nil, err
		}
		user.Password = ""
		return user, nil
	}
}

// Get returns a user from the DB
func Get(u UserDB) func(ctx context.Context, username string) (interface{}, error) {
	return func(ctx context.Context, username string) (interface{}, error) {
		return u.GetUser(ctx, username)
	}
}

// IsPasswordCorrect checks if the username and password match the ones in the DB
func IsPasswordCorrect(u UserDB) func(ctx context.Context, username, password string) error {
	return func(ctx context.Context, username, password string) error {
		hashedPassword, err := u.GetPasswordForUser(ctx, username)
		if err != nil {
			return err
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		return err
	}
}

func hashPasswordForUser(u *User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not generate hash for user '%s'", u.Username)
	}
	u.Password = Password(hash)
	return nil
}
