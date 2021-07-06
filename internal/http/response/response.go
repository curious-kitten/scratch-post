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
func SendError(w http.ResponseWriter, message string, code int) error {
	execError := &executionError{
		Error: message,
		Code:  code,
	}

	js, err := json.Marshal(execError)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(js)
	return err
}

// Send writes a json body to the response writter
func Send(w http.ResponseWriter, value interface{}, code int) error{
	js, err := json.Marshal(value)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(js)
	return err
}
