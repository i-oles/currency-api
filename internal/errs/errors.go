package errs

import "net/http"

const (
	statusCode400 = http.StatusBadRequest
)

type AppError struct {
	Message string `json:"message"`
	Code    int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func New(msg string, code int, err error) *AppError {
	return &AppError{
		Message: msg,
		Code:    code,
		Err:     err,
	}
}

func NotFound(msg string, err error) *AppError {
	return New(msg, http.StatusNotFound, err)
}

func APIResponseError(msg string, err error) *AppError {
	return New(msg, statusCode400, err)
}
