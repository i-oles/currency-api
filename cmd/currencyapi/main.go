package main

import (
	"log/slog"
	openExchange "main/internal/api/openexchange"
	"main/internal/handlers/rates"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/")

	openExchangeAPI := openExchange.New()

	ratesHandler := rates.NewHandler(openExchangeAPI)
	api.GET("/rates", ratesHandler.Handle)

	slog.Info("Starting server...", slog.String("listen address", ":8080"))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error("listenAndServe", slog.String("err", err.Error()))
	}
}
