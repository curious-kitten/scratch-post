package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/http/response"
)

type checkPassword func(ctx context.Context, username, password string) error

// LoginRequest is used to request access to the API
type LoginRequest struct {
	Username string
	Password string
	Duration int
}

// Validate the Login Request
func (u *LoginRequest) Validate() error {
	if u.Username == "" {
		return decoder.NewValidationError("username not provided")
	}
	if u.Password == "" {
		return decoder.NewValidationError("password not provided")
	}
	return nil
}

// Endpoints used to authorize an end user
type Endpoints struct {
	token             Authorizer
	isPasswordCorrect checkPassword
	ctx               context.Context
}

// Authorizer is used to validate the user request
type Authorizer interface {
	GenerateSecurityString(username string) (string, time.Time, error)
	Invalidate(token string) error
	Validate(token string) (bool, string, error)
	Cleanup(cleanInterval time.Duration)
}

// NewEndpoints is used setup the autorization endpoints
func NewEndpoints(ctx context.Context, isPasswordCorrect checkPassword, token Authorizer) *Endpoints {
	return &Endpoints{
		token:             token,
		isPasswordCorrect: isPasswordCorrect,
		ctx:               ctx,
	}
}

func (e *Endpoints) login(w http.ResponseWriter, r *http.Request) {
	user := &LoginRequest{}
	if err := decoder.Decode(user, r.Body); err != nil {
		response.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	toctx, cancel := context.WithTimeout(e.ctx, time.Second*10)
	defer cancel()
	err := e.isPasswordCorrect(toctx, user.Username, user.Password)
	if err != nil {
		response.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tokenString, expirationTime, err := e.token.GenerateSecurityString(user.Username)

	if err != nil {
		response.SendError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	SetAuthCookie(w, tokenString, expirationTime)
}

func (e *Endpoints) logout(w http.ResponseWriter, r *http.Request) {
	token, err := GetToken(r)
	if err != nil {
		if err == http.ErrNoCookie {
			response.SendError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = e.token.Invalidate(token)
	if err != nil {
		response.SendError(w, err.Error(), http.StatusInternalServerError)
	}
}

// Register is used to attach the authorization endpoints to a given mux router
func (e *Endpoints) Register(r *mux.Router) {
	r.HandleFunc("/login", e.login).Methods(http.MethodPost)
	r.HandleFunc("/logout", e.logout).Methods(http.MethodPost)
}

// SetAuthCookie add an auth cookie to the response writer
func SetAuthCookie(w http.ResponseWriter, tokenString string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:    authToken,
		Value:   tokenString,
		Expires: expires,
	})
}

// GetAuthToken extracts the auth token from a request
func GetToken(r *http.Request) (string, error) {
	c, err := r.Cookie(authToken)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// AddUserIDHeader adds a header to the request that represents the user ID
func AddUserIDHeader(r *http.Request, username string) {
	r.Header.Add(userIDKey, fmt.Sprint(username))
}

// RemoveUserIDHeader removes the header that contains the user ID
func RemoveUserIDHeader(r *http.Request) {
	r.Header.Del(userIDKey)
}

// GetUserIDFromRequest returns the user ID from the request
func GetUserIDFromRequest(r *http.Request) (string, error) {
	id := r.Header.Get(userIDKey)
	if id == "" {
		return "", fmt.Errorf("request is missing user ID")
	}
	return id, nil
}
