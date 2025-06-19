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

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx, currencies)
	if err != nil {
		//TODO: add note about the fact that it is not a badRequest
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	currencyCombinations, err := getAllCombinations(currencies)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	result, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

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

func getAllCombinations(input []string) ([][]string, error) {
	n := len(input)
	if n < 2 {
		return nil, errors.New("not enough elements in to get the combinations")
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

	if currencyCombinations == nil || len(currencyCombinations) == 0 {
		return nil, errors.New("no combinations")
	}

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

func roundFloat(val float64, places int) float64 {
	factor := math.Pow(10, float64(places))

	return math.Round(val*factor) / factor
}
