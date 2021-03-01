package middleware

import (
	"net/http"
	"runtime/debug"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/curious-kitten/scratch-post/pkg/http/auth"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

type logger interface {
	Errorw(msg string, keysAndValues ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
}

// Logging wraprs a handler to perform logging when requests are made
func Logging(logger logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					stack := string(debug.Stack())
					logger.Errorw(
						"error occurred",
						"err", err,
						"trace", stack,
					)
				}
			}()

			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			logger.Debugw(
				"request received",
				"status", wrapped.Status(),
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"query", r.URL.Query(),
			)
		}

		return http.HandlerFunc(fn)
	}
}

// Authorization verifies a request has a valid token associated
func Authorization(authorizer auth.Authorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetAuthToken(r)
			if err != nil {
				if err == http.ErrNoCookie {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			valid, username, err := authorizer.Validate(token)
			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			auth.AddUserIDHeader(r, username)
			defer auth.RemoveUserIDHeader(r)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
