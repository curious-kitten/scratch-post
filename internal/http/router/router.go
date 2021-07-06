package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/http/response"
	"github.com/curious-kitten/scratch-post/internal/logger"
)

// New creates a new mux router
func New(log logger.Logger) *mux.Router {
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = methodNotAllowedHandler()
	r.NotFoundHandler = notFoundHandler()
	r.Use(logger.HTTPLogging(log))
	return r
}

func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.SendError(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}

func notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.SendError(w, "not found", http.StatusNotFound)
	})
}
