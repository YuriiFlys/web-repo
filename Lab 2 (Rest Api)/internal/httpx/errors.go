package httpx

import "net/http"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Err(code, msg string) APIError {
	return APIError{Code: code, Message: msg}
}

const (
	CodeBadRequest = "BAD_REQUEST"
	CodeNotFound   = "NOT_FOUND"
)

func StatusFor(code string) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
