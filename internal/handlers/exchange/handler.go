package exchange

import (
	"encoding/json"
	"fmt"
	"main/internal/errs"
	"main/internal/repository"
	"main/internal/repository/memory"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Response struct {
	From   string      `json:"from"`
	To     string      `json:"to"`
	Amount json.Number `json:"amount"`
}

type Handler struct {
	currencyRateRepo repository.CurrencyRate
	errorHandler     errs.ErrorHandler
}

func NewHandler(
	currencyRateRepo repository.CurrencyRate,
	errorHandler errs.ErrorHandler,
) *Handler {
	return &Handler{
		currencyRateRepo: currencyRateRepo,
		errorHandler:     errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	if err := ctx.Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service is shutting down"})

		return
	}

	resp, err := h.exchange(c)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) exchange(c *gin.Context) (Response, error) {
	sourceCurrency := c.Query("from")
	targetCurrency := c.Query("to")
	amountStr := c.Query("amount")

	if sourceCurrency == "" || targetCurrency == "" || amountStr == "" {
		return Response{}, errs.ErrEmptyParam
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return Response{}, errs.ErrAmountNotNumber
	}

	if amount.LessThan(decimal.NewFromFloat(0)) {
		return Response{}, errs.ErrNegativeAmount
	}

	sourceCurrencyDetails, err := h.currencyRateRepo.Get(sourceCurrency)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get source currency rate: %w", err)
	}

	targetCurrencyDetails, err := h.currencyRateRepo.Get(targetCurrency)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get the target currency rate: %w", err)
	}

	if zeroValue(sourceCurrencyDetails, targetCurrencyDetails) {
		return Response{}, errs.ErrZeroValue
	}

	exchangeResult, err := calculateExchange(sourceCurrencyDetails, targetCurrencyDetails, amount)
	if err != nil {
		return Response{}, fmt.Errorf("failed to calculate exchange rate: %w", err)
	}

	return Response{
		From:   sourceCurrency,
		To:     targetCurrency,
		Amount: json.Number(exchangeResult),
	}, nil
}

func calculateExchange(
	sourceCurrencyDetails,
	targetCurrencyDetails memory.CurrencyDetails,
	amount decimal.Decimal,
) (string, error) {
	sourceRate := decimal.NewFromFloat(sourceCurrencyDetails.Rate)

	targetRate := decimal.NewFromFloat(targetCurrencyDetails.Rate)

	exchangeRate := sourceRate.Div(targetRate)

	result := amount.Mul(exchangeRate)

	decimalPlaces := int32(targetCurrencyDetails.DecimalPrecision)

	return result.StringFixed(decimalPlaces), nil
}

func zeroValue(source, target memory.CurrencyDetails) bool {
	return source.Rate == 0 || target.Rate == 0 || source.DecimalPrecision == 0 || target.DecimalPrecision == 0
}
