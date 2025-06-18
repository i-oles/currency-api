package rates

import (
	"errors"
	"fmt"
	"main/internal/api"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	currencyRateAPI api.CurrencyRate
}

func NewHandler(currencyRateAPI api.CurrencyRate) *Handler {
	return &Handler{
		currencyRateAPI: currencyRateAPI,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	param := c.Query("currencies")
	err := validateParameter(param)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	currencies := strings.Split(param, ",")

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx)
	if err != nil {
		//TODO: add note about the fact that it is not a badRequest
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	currencyCombinations := getAllCombinations(currencies)

	result, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, result)
}

func validateParameter(param string) error {
	if param == "" {
		return errors.New("param cannot be empty")
	}

	currencies := strings.Split(param, ",")

	if len(currencies) < 2 {
		return errors.New("not enough param given")
	}

	return nil
}

func calculateCurrencyRates(
	rates map[string]float64,
	currencyCombinations [][]string,
) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, combination := range currencyCombinations {
		if len(combination) != 2 {
			return nil, errors.New("one combination should contain only two values")
		}

		currency1 := combination[0]
		currency2 := combination[1]

		rate1, ok := rates[currency1]
		if !ok {
			return nil, fmt.Errorf("api do not provide %s rate", currency1)
		}

		rate2, ok := rates[currency2]
		if !ok {
			return nil, fmt.Errorf("api do not provide %s rate", currency2)
		}

		result = append(result, map[string]interface{}{
			"from": currency1,
			"to":   currency2,
			"rate": roundFloat(rate2/rate1, 6),
		})
	}

	return result, nil
}

func getAllCombinations(input []string) [][]string {
	var result [][]string
	for i := 0; i < len(input); i++ {
		for j := 0; j < len(input); j++ {
			if i != j {
				result = append(result, []string{input[i], input[j]})
			}
		}
	}
	return result
}

func roundFloat(val float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(val*factor) / factor
}
