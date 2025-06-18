package rates

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
