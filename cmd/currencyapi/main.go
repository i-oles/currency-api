package main

import (
	"main/internal/handlers/rates"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/")

	ratesHandler := rates.NewHandler()
	api.GET("/rates", ratesHandler.Handle)
}
