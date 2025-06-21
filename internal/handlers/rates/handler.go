package rates

import (
	"context"
	"errors"
	"fmt"
	"main/internal/api"
	"main/internal/errs"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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

	param := c.Query("currencies")
	if param == "" {
		c.AbortWithStatus(http.StatusBadRequest)

		return
	}

	result, err := h.countRates(ctx, param)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) countRates(ctx context.Context, param string) ([]map[string]interface{}, error) {
	currencies := strings.Split(param, ",")
	if len(currencies) < 2 {
		return nil, errs.ErrBadRequest
	}

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx, currencies)
	if err != nil {
		return nil, err
	}

	currencyCombinations, err := getAllCombinations(currencies)
	if err != nil {
		return nil, fmt.Errorf("failed to get combinations: %w", err)
	}

	result, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getAllCombinations(input []string) ([][]string, error) {
	n := len(input)
	if n < 2 {
		return nil, errors.New("not enough elements to generate combinations")
	}

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

func calculateCurrencyRates(
	rates map[string]float64,
	currencyCombinations [][]string,
) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

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

		targetRate, ok := rates[targetCurrency]
		if !ok {
			return nil, errs.ErrCurrencyNotFound
		}

		result = append(result, map[string]interface{}{
			"from": sourceCurrency,
			"to":   targetCurrency,
			"rate": roundFloat(targetRate/sourceRate, 6),
		})
	}

	return result, nil
}

func roundFloat(val float64, places int) float64 {
	factor := math.Pow(10, float64(places))

	return math.Round(val*factor) / factor
}
