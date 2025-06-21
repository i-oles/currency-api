package exchange

import (
	"errors"
	"fmt"
	"main/internal/errs"
	"main/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

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

	sourceCurrency := c.Query("from")
	targetCurrency := c.Query("to")
	amountStr := c.Query("amount")

	if sourceCurrency == "" || targetCurrency == "" || amountStr == "" {
		c.AbortWithStatus(http.StatusBadRequest)

		return
	}

	result, err := h.exchange(sourceCurrency, targetCurrency, amountStr)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	fmt.Printf("result string: %+v\n", result)

	jsonData := fmt.Sprintf(
		`{"from":"%s","to":"%s","amount":%s}`,
		sourceCurrency,
		targetCurrency,
		result,
	)

	fmt.Println(jsonData)

	c.Data(http.StatusOK, "application/json", []byte(jsonData))
}

func (h *Handler) exchange(
	sourceCurrency, targetCurrency, amountStr string,
) (string, error) {

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return "", errs.ErrBadRequest
	}

	if amount.LessThan(decimal.NewFromFloat(0)) {
		return "", errs.ErrBadRequest
	}

	sourceCurrencyDetails, err := h.currencyRateRepo.Get(sourceCurrency)
	if err != nil {
		return "", err
	}

	targetCurrencyDetails, err := h.currencyRateRepo.Get(targetCurrency)
	if err != nil {
		return "", err
	}

	if len(sourceCurrencyDetails) != 2 && len(targetCurrencyDetails) != 2 {
		return "", errors.New("len of currency details is invalid")
	}

	sourceRate := decimal.NewFromFloat(sourceCurrencyDetails[1])
	targetRate := decimal.NewFromFloat(targetCurrencyDetails[1])

	exchangeRate := sourceRate.Div(targetRate)

	result := amount.Mul(exchangeRate)

	decimalPlaces := int32(targetCurrencyDetails[0])

	return result.StringFixed(decimalPlaces), nil
}
