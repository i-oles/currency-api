package main

import (
	"context"
	"errors"
	"log/slog"
	openExchange "main/internal/api/openexchange"
	"main/internal/configuration"
	"main/internal/errs"
	"main/internal/errs/currency"
	logging "main/internal/errs/log"
	"main/internal/handlers/exchange"
	"main/internal/handlers/rates"
	"main/internal/repository/memory"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	_ = godotenv.Load()

	appID := os.Getenv("APP_ID")
	if appID == "" {
		slog.Error(
			"personal APP_ID for openExchangeAPI is not set." +
				"Please set APP_ID env. Details in the README.md",
		)
		os.Exit(1)
	}

	slog.Info("appID for openExchangeAPI:", slog.String("appID", appID))

	var cfg configuration.Configuration

	err := configuration.GetConfig("./config", &cfg)
	if err != nil {
		slog.Error("Error loading configuration:", slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	router := gin.Default()

	api := router.Group("/")

	var errorHandler errs.ErrorHandler

	errorHandler = currency.NewErrorHandler()
	if cfg.LogErrors {
		errorHandler = logging.NewErrorHandler(errorHandler)
	}

	openExchangeAPI := openExchange.New(cfg.APIURL, appID)

	ratesHandler := rates.NewHandler(openExchangeAPI, errorHandler)
	api.GET("/rates", ratesHandler.Handle)

	currencyRateRepo := memory.NewCurrencyRateRepo()
	exchangeHandler := exchange.NewHandler(currencyRateRepo, errorHandler)

	api.GET("/exchange", exchangeHandler.Handle)

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
			slog.Error("ListenAndServe error: %s\n", slog.String("err", err.Error()))
			os.Exit(1)
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
