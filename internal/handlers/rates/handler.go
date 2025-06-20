package rates

import (
	"context"
	"errors"
	"main/internal/api"
	"main/internal/errs"
	"math"
	"net/http"
	"strings"
	"time"

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
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	currencies := strings.Split(param, ",")
	if len(currencies) < 2 {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	apiCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := h.currencyRateAPI.GetCurrencyRates(apiCtx, currencies)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	currencyCombinations, err := getAllCombinations(currencies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	result, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		h.errorHandler.Handle(c, err)

		return
	}

	c.JSON(http.StatusOK, result)
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

		currency1 := combination[0]
		currency2 := combination[1]

		rate1, ok := rates[currency1]
		if !ok {
			return nil, errs.ErrCurrencyNotFound
		}

		rate2, ok := rates[currency2]
		if !ok {
			return nil, errs.ErrCurrencyNotFound
		}

		result = append(result, map[string]interface{}{
			"from": currency1,
			"to":   currency2,
			"rate": roundFloat(rate2/rate1, 6),
		})
	}

	return result, nil
}

func roundFloat(val float64, places int) float64 {
	factor := math.Pow(10, float64(places))

	return math.Round(val*factor) / factor
}
