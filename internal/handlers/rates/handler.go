package rates

import (
	"main/internal/api"
	"net/http"

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
	c.JSON(http.StatusOK, gin.H{})
}
