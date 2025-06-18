package main

import (
	openExchange "main/internal/api/openexchange"
	"main/internal/handlers/rates"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/")

	openExchangeAPI := openExchange.New()

	ratesHandler := rates.NewHandler(openExchangeAPI)
	api.GET("/rates", ratesHandler.Handle)
}
