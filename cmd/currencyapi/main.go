package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	openExchange "main/internal/api/openexchange"
	"main/internal/configuration"
	"main/internal/handlers/rates"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	var cfg configuration.Configuration

	err := configuration.GetConfig("./config", &cfg)
	if err != nil {
		slog.Error(err.Error())
	}

	log.Println(cfg.Pretty())

	router := gin.Default()

	api := router.Group("/")

	openExchangeAPI := openExchange.New(cfg.APIURL)

	ratesHandler := rates.NewHandler(openExchangeAPI)
	api.GET("/rates", ratesHandler.Handle)

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout * time.Second,
		ReadTimeout:       cfg.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WriteTimeout * time.Second,
	}

	go func() {
		slog.Info("Starting server...", slog.String("listen address", cfg.ListenAddress))

		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe error: %s\n", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...", slog.String("listen address", cfg.ListenAddress))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ContextTimeout*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown:", slog.String("err", err.Error()))
	}

	slog.Info("Server exiting...", slog.String("listen address", cfg.ListenAddress))
}
