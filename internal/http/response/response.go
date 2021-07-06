package response

import (
	"encoding/json"
	"net/http"
)

type executionError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// SendError returns an error message with the given status code
func SendError(w http.ResponseWriter, message string, code int) {
	execError := &executionError{
		Error: message,
		Code:  code,
	}

	js, err := json.Marshal(execError)
	if err != nil {
		SendError(w, err.Error(), 400)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(js)
}

// Send writes a json body to the response writter
func Send(w http.ResponseWriter, value interface{}, code int) {
	js, err := json.Marshal(value)
	if err != nil {
		SendError(w, err.Error(), 400)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(js)
}
