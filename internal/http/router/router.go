package router

import (
	"net/http"

	"github.com/curious-kitten/scratch-post/internal/http/helpers"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/http/middleware"
	"github.com/curious-kitten/scratch-post/internal/logger"
)

// New creates a new mux router
func New(log logger.Logger) *mux.Router {
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = methodNotAllowedHandler()
	r.NotFoundHandler = notFoundHandler()
	r.Use(middleware.Logging(log))
	return r
}

func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.FormatError(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}

func notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.FormatError(w, "not found", http.StatusNotFound)
	})
}
