package probes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/http/helpers"
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
			helpers.FormatResponse(w, status, http.StatusOK)
		} else {
			helpers.FormatResponse(w, status, http.StatusInternalServerError)
		}
	}
}
