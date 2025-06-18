package rates

import (
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

func (h *Handler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	resp, err := h.currencyRateAPI.GetCurrencyRates(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	param := c.Query("currencies")

	currencies := strings.Split(param, ",")

	rates := make(map[string]float64)

	for _, cur := range currencies {
		val, ok := resp.Rates[cur]
		if !ok {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("api do not provide %s rate", cur)},
			)
		}

		rates[cur] = val
	}

	fmt.Printf("currencies: %+v", rates)

	c.JSON(http.StatusOK, gin.H{})
}
