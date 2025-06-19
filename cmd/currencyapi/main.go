package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	openExchange "main/internal/api/openexchange"
	"main/internal/handlers/rates"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/")

	openExchangeAPI := openExchange.New()

	ratesHandler := rates.NewHandler(openExchangeAPI)
	api.GET("/rates", ratesHandler.Handle)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		slog.Info("Starting server...", slog.String("listen address", ":8080"))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe error: %s\n", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...", slog.String("listen address", ":8080"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", slog.String("err", err.Error()))
	}

	slog.Info("Server exiting...", slog.String("listen address", ":8080"))
}
