package rates

import (
	"errors"
	"fmt"
	"main/internal/api"
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

	//TODO: add validation request params ("", missing parameter, one, parameter --> error)
	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	param := c.Query("currencies")
	currencies := strings.Split(param, ",")

	currencyCombinations := getPermutations(currencies)

	r, err := calculateCurrencyRates(resp.Rates, currencyCombinations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, r)
}

func calculateCurrencyRates(
	rates map[string]float64,
	currencyCombinations [][]string,
) ([]map[string]interface{}, error) {
	baseCur := "USD"

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

			r = append(r, map[string]interface{}{"from": baseCur, "to": combination[1], "rate": val})

			continue
		}

		if combination[1] == baseCur {
			val, ok := rates[combination[0]]
			if !ok {
				return nil, fmt.Errorf("api do not provide %s rate", combination[0])
			}

			r = append(r, map[string]interface{}{
				"from": combination[0],
				"to":   baseCur,
				"rate": 1 / val,
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

		r = append(r, map[string]interface{}{"from": combination[0], "to": combination[1], "rate": val2 / val1})

	}

	return r, nil
}

func getPermutations(input []string) [][]string {
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
