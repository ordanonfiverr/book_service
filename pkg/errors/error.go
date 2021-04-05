package errors

import "fmt"

type HttpError struct {
	Code int
	Message string
	Err error
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("failed with error: %s, http code: %d, inner err: %s", e.Message, e.Code, e.Err)
}

var _ error = &HttpError{}

func NewHttpError(code int, message string, err error) *HttpError {
	return &HttpError{
		Code: code,
		Message: message,
		Err: err,
	}
}
