package middleware

import (
	"net/http"
	"runtime/debug"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/curious-kitten/scratch-post/internal/http/authorization"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        []byte
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

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body
	return rw.ResponseWriter.Write(body)
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
				"body", string(wrapped.body),
			)
		}

		return http.HandlerFunc(fn)
	}
}

// JWTAuthorization verifies a request has a valid JWT token associated
func JWTAuthorization(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		token, err := authorization.GetAuthToken(r)
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		valid, claim, err := authorization.ValidateToken(token)
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

		ok := authorization.IsBlacklisted(token)
		if ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authorization.AddUserIDHeader(r, claim)
		defer authorization.RemoveUserIDHeader(r)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
