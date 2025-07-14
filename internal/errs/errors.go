package errs

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrAPIResponse          = errors.New("error api response error")
	ErrRepoCurrencyNotFound = errors.New("error repo currency not found error")
	ErrCurrencyNotFound     = errors.New("error unknown currency")
	ErrBadRequest           = errors.New("error invalid request")
	ErrNegativeAmount       = errors.New("error amount must be positive number")
	ErrEmptyParam           = errors.New("error one or more params is empty")
	ErrAmountNotNumber      = errors.New("error amount must a number")
	ErrZeroValue            = errors.New("error got zero value from API or Repository")
)

type ErrorHandler interface {
	Handle(c *gin.Context, err error)
}
