package helpers

import (
	"encoding/json"
	"net/http"
)

type executionError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// FormatError returns an error message with the given status code
func FormatError(w http.ResponseWriter, message string, code int) {
	execError := &executionError{
		Error: message,
		Code:  code,
	}

	js, err := json.Marshal(execError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(js)
}

// FormatResponse writes a json body to the response writter
func FormatResponse(w http.ResponseWriter, value interface{}, code int) {
	js, err := json.Marshal(value)
	if err != nil {
		FormatError(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(js)
}
