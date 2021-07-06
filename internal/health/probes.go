package health

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/http/response"
)

type check func() (bool, interface{})

type conditions interface {
	IsReady() (bool, interface{})
	IsAlive() (bool, interface{})
}

// RegisterHTTPProbes adds the probe api to the given router
func RegisterHTTPProbes(r *mux.Router, c conditions) {
	r.HandleFunc("/alive", conditionHandler(c.IsAlive)).Methods(http.MethodGet)
	r.HandleFunc("/ready", conditionHandler(c.IsReady)).Methods(http.MethodGet)
}

func conditionHandler(c check) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ready, status := c()
		if ready {
			response.Send(w, status, http.StatusOK)
		} else {
			response.Send(w, status, http.StatusInternalServerError)
		}
	}
}
