package exchange

import (
	"context"
	"errors"
	"main/internal/errs"
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Handler struct {
	CurrencyRateRepo repository.CurrencyRate
}

func NewHandler(currencyRateRepo repository.CurrencyRate) *Handler {
	return &Handler{
		CurrencyRateRepo: currencyRateRepo,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	if err := ctx.Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service is shutting down"})

		return
	}

	sourceCurrency := c.Query("from")
	targetCurrency := c.Query("to")
	amountStr := c.Query("amount")

	if sourceCurrency == "" || targetCurrency == "" || amountStr == "" {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error Decimal:": err.Error()})

		return
	}

	if amount.LessThan(decimal.NewFromFloat(0)) {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	sourceCurrencyDetails, err := h.CurrencyRateRepo.Get(sourceCurrency)
	if err != nil {
		h.handleCurrencyRateError(c, err)

		return
	}

	targetCurrencyDetails, err := h.CurrencyRateRepo.Get(targetCurrency)
	if err != nil {
		h.handleCurrencyRateError(c, err)

		return
	}

	if len(sourceCurrencyDetails) != 2 && len(targetCurrencyDetails) != 2 {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	sourceRate := decimal.NewFromFloat(sourceCurrencyDetails[1])
	targetRate := decimal.NewFromFloat(targetCurrencyDetails[1])

	exchangeRate := sourceRate.Div(targetRate)

	result := amount.Mul(exchangeRate)

	c.JSON(http.StatusOK, gin.H{
		"from":   sourceCurrency,
		"to":     targetCurrency,
		"amount": result.StringFixed(int32(targetCurrencyDetails[0])),
	})
}

func (h *Handler) handleCurrencyRateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		h.sendErrorResponse(c, http.StatusGatewayTimeout, "currency rate API timeout")
	case errors.Is(err, errs.ErrAPIResponse):
		h.sendErrorResponse(c, errs.StatusCode400, "")
	case errors.Is(err, errs.ErrCurrencyNotFound):
		h.sendErrorResponse(c, http.StatusNotFound, errs.ErrCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrRepoCurrencyNotFound):
		h.sendErrorResponse(c, http.StatusBadRequest, errs.ErrRepoCurrencyNotFound.Error())
	default:
		h.sendErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
}

func (h *Handler) sendErrorResponse(c *gin.Context, status int, message string) {
	if message == "" {
		c.JSON(status, nil)
	} else {
		c.JSON(status, gin.H{"error": message})
	}
}
