package errs

import (
	"fmt"
)

// JSONError represents different types of JSON-related errors
type JSONError struct {
	Message string  `json:"message"`
	Input   string  `json:"input"`
	Err     error   `json:"err"`
	Field   *string `json:"field"`
}

func (e *JSONError) Error() string {
	return e.Message
}

func (e *JSONError) Unwrap() error {
	return e.Err
}

type UnexpectedError struct {
	Message string
	Err     error
}

func (e *UnexpectedError) Error() string {
	return e.Message
}

func (e *UnexpectedError) Unwrap() error {
	return e.Err
}

type ApiError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Details    any    `json:"details"`
}

func (e ApiError) Error() string {
	return e.Message
}

func NewApiError(message string, statusCode int, details any) ApiError {
	return ApiError{
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

func InvalidJSON() ApiError {
	return NewApiError("invalid JSON request data", 400, nil)
}

func InvalidRequestData(errors map[string]string) ApiError {
	return ApiError{
		StatusCode: 400,
		Details:    errors,
	}
}

func MissingControllerArgsError(args any) error {
	return fmt.Errorf("missing controller arguments %+v", args)
}
