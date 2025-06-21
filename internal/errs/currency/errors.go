package currency

import (
	"context"
	"errors"
	"main/internal/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (e *ErrorHandler) Handle(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		e.sendErrorResponse(c, http.StatusGatewayTimeout, "currency rate API timeout")
	case errors.Is(err, errs.ErrCurrencyNotFound):
		e.sendErrorResponse(c, http.StatusNotFound, errs.ErrCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrAPIResponse),
		errors.Is(err, errs.ErrRepoCurrencyNotFound),
		errors.Is(err, errs.ErrNegativeAmount),
		errors.Is(err, errs.ErrAmountNotNumber),
		errors.Is(err, errs.ErrEmptyParam),
		errors.Is(err, errs.ErrBadRequest):
		e.sendErrorResponse(c, http.StatusBadRequest, "")
	default:
		e.sendErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
}

func (e *ErrorHandler) sendErrorResponse(c *gin.Context, status int, message string) {
	if message == "" {
		c.AbortWithStatus(status)
	} else {
		c.JSON(status, gin.H{"error": message})
	}
}
