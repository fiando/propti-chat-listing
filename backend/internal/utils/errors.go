package utils

import (
	"encoding/json"
	"fmt"
)

// AppError is a structured error with an HTTP status code and message.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrNotFound     = &AppError{Code: 404, Message: "Resource not found"}
	ErrUnauthorized = &AppError{Code: 401, Message: "Unauthorized"}
	ErrForbidden    = &AppError{Code: 403, Message: "Forbidden"}
	ErrBadRequest   = &AppError{Code: 400, Message: "Bad request"}
	ErrInternal     = &AppError{Code: 500, Message: "Internal server error"}
)

// NewAppError creates a custom AppError with the given HTTP code and message.
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WrapError wraps an underlying error into an AppError.
func WrapError(appErr *AppError, err error) *AppError {
	return &AppError{Code: appErr.Code, Message: fmt.Sprintf("%s: %s", appErr.Message, err.Error())}
}

// ErrorResponse is the JSON body returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// MarshalErrorResponse serialises an AppError into a JSON string for API responses.
func MarshalErrorResponse(appErr *AppError) string {
	resp := ErrorResponse{Error: appErr.Message, Code: appErr.Code}
	b, _ := json.Marshal(resp)
	return string(b)
}
