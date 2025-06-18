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

type Curr struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	param := c.Query("currencies")
	currencies := strings.Split(param, ",")

	err := validateParameter(currencies)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
	}

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)

		return
	}

	currencyCombinations := getAllCombinations(currencies)

	result, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, result)
}

func validateParameter(currencies []string) error {
	if len(currencies) < 2 {
		return errors.New("not enough currencies given")
	}

	return nil
}

func calculateCurrencyRates(
	rates map[string]float64,
	currencyCombinations [][]string,
) ([]map[string]interface{}, error) {
	baseCur := "USD"
	baseCurRate := 1.0

	r := make([]map[string]interface{}, 0)

	for _, combination := range currencyCombinations {
		if len(combination) != 2 {
			return nil, errors.New("one combination should contain only two values")
		}

		if combination[0] == baseCur {
			val, ok := rates[combination[1]]
			if !ok {
				return nil, fmt.Errorf("api do not provide %s rate", combination[1])
			}

			r = append(r, map[string]interface{}{
				"from": baseCur,
				"to":   combination[1],
				"rate": val,
			})

			continue
		}

		if combination[1] == baseCur {
			val, ok := rates[combination[0]]
			if !ok {
				return nil, fmt.Errorf("api do not provide %s rate", combination[0])
			}

			div := baseCurRate / val

			r = append(r, map[string]interface{}{
				"from": combination[0],
				"to":   baseCur,
				"rate": roundFloat(div, 6),
			})

			continue
		}

		val1, ok := rates[combination[0]]
		if !ok {
			return nil, fmt.Errorf("api do not provide %s rate", combination[0])
		}

		val2, ok := rates[combination[1]]
		if !ok {
			return nil, fmt.Errorf("api do not provide %s rate", combination[1])
		}

		div := val2 / val1

		r = append(r, map[string]interface{}{
			"from": combination[0],
			"to":   combination[1],
			"rate": roundFloat(div, 6),
		})
	}

	return r, nil
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
