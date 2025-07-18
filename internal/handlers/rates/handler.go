package rates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/internal/api"
	"main/internal/errs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	decimalPrecision = 8
)

type Response struct {
	From string      `json:"from"`
	To   string      `json:"to"`
	Rate json.Number `json:"rate"`
}

type Handler struct {
	currencyRateAPI api.CurrencyRate
	errorHandler    errs.ErrorHandler
}

func NewHandler(
	currencyRateAPI api.CurrencyRate,
	errorHandler errs.ErrorHandler,
) *Handler {
	return &Handler{
		currencyRateAPI: currencyRateAPI,
		errorHandler:    errorHandler,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	if err := ctx.Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service is shutting down"})

		return
	}

	responses, err := h.countRates(ctx, c)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, responses)
}

func (h *Handler) countRates(ctx context.Context, c *gin.Context) ([]Response, error) {
	param := c.Query("currencies")
	if param == "" {
		return nil, errs.ErrEmptyParam
	}

	currencies := strings.Split(param, ",")
	if len(currencies) < 2 || containsDuplicates(currencies) {
		return nil, errs.ErrBadRequest
	}

	currencyCombinations, err := getAllCombinations(currencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get combinations: %w", err)
	}

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx, currencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get currency rates: %w", err)
	}

	responses, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func getAllCombinations(input []string) ([][]string, error) {
	n := len(input)

	result := make([][]string, 0, n*(n-1))

	for _, curr1 := range input {
		for _, curr2 := range input {
			if curr1 != curr2 {
				result = append(result, []string{curr1, curr2})
			}
		}
	}

	return result, nil
}

func containsDuplicates(list []string) bool {
	seen := make(map[string]bool)
	for _, item := range list {
		if ok := seen[item]; ok {
			return true
		}
		seen[item] = true
	}

	return false
}

func calculateCurrencyRates(
	rates map[string]float64,
	currencyCombinations [][]string,
) ([]Response, error) {
	responses := make([]Response, 0, len(currencyCombinations))

	if len(currencyCombinations) == 0 {
		return nil, errors.New("no combinations given")
	}

	for _, combination := range currencyCombinations {
		if len(combination) != 2 {
			return nil, errors.New("one combination should contain exactly two values")
		}

		sourceCurrency := combination[0]
		targetCurrency := combination[1]

		sourceRate, ok := rates[sourceCurrency]
		if !ok {
			return nil, errs.ErrCurrencyNotFound
		}

		if sourceRate == 0 {
			return nil, errs.ErrZeroValue
		}

		targetRate, ok := rates[targetCurrency]
		if !ok {
			return nil, errs.ErrCurrencyNotFound
		}

		sourceDecimalRate := decimal.NewFromFloat(sourceRate)
		targetDecimalRate := decimal.NewFromFloat(targetRate)

		decimalRate := targetDecimalRate.Div(sourceDecimalRate)

		strRate := decimalRate.StringFixed(decimalPrecision)
		response := Response{
			From: sourceCurrency,
			To:   targetCurrency,
			Rate: json.Number(strRate),
		}

		responses = append(responses, response)
	}

	return responses, nil
}
