package errs

import (
	"errors"
	"net/http"
)

const (
	StatusCode400 = http.StatusBadRequest
)

var (
	ErrAPIResponse      = errors.New("error api response error")
	ErrCurrencyNotFound = errors.New("error unknown currency")
)
