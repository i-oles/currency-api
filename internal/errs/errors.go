package errs

import (
	"errors"
	"net/http"
)

const (
	StatusCode400 = http.StatusBadRequest
)

var (
	ErrAPIResponse          = errors.New("error api response error")
	ErrRepoCurrencyNotFound = errors.New("error repo currency not found error")
	ErrCurrencyNotFound     = errors.New("error unknown currency")
	ErrBadRequest           = errors.New("error invalid request")
)
