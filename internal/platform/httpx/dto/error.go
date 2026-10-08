package dto

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Error{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
