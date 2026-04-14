package utils

import "fmt"

type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

var (
	NotFound     = func(msg string) *AppError { return &AppError{Code: "NOT_FOUND", Message: msg, Status: 404} }
	Unauthorized = func(msg string) *AppError { return &AppError{Code: "UNAUTHORIZED", Message: msg, Status: 401} }
	BadRequest   = func(msg string) *AppError { return &AppError{Code: "BAD_REQUEST", Message: msg, Status: 400} }
	Conflict     = func(msg string) *AppError { return &AppError{Code: "CONFLICT", Message: msg, Status: 409} }
	InternalErr  = func(msg string) *AppError { return &AppError{Code: "INTERNAL_ERROR", Message: msg, Status: 500} }
)

func WrapError(code, msg string, err error, status int) *AppError {
	return &AppError{Code: code, Message: msg, Err: err, Status: status}
}
