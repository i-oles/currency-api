package main

import (
	"context"
	"errors"
	"fmt"
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

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	router := setupRouter(cfg)

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadTimeout * time.Second,
		ReadTimeout:       cfg.ReadTimeout * time.Second,
		WriteTimeout:      cfg.WriteTimeout * time.Second,
	}

	runServer(srv, cfg)
}

func runServer(srv *http.Server, cfg configuration.Configuration) {
	go func() {
		slog.Info("Starting server...", slog.String("address", cfg.ListenAddress))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error: %s\n", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ContextTimeout*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("err", err.Error()))
	}

	slog.Info("Server stopped")
}

func setupRouter(cfg configuration.Configuration) *gin.Engine {
	router := gin.Default()

	api := router.Group("/")

	var errorHandler errs.ErrorHandler

	errorHandler = currency.NewErrorHandler()
	if cfg.LogErrors {
		errorHandler = logging.NewErrorHandler(errorHandler)
	}

	openExchangeAPI := openExchange.New(cfg.APIURL, os.Getenv("APP_ID"))

	ratesHandler := rates.NewHandler(openExchangeAPI, errorHandler)
	api.GET("/rates", ratesHandler.Handle)

	currencyRateRepo := memory.NewCurrencyRateRepo()
	exchangeHandler := exchange.NewHandler(currencyRateRepo, errorHandler)

	api.GET("/exchange", exchangeHandler.Handle)

	return router
}

func loadConfig() (configuration.Configuration, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using environment variables...")
	}

	appID := os.Getenv("APP_ID")
	if appID == "" {
		return configuration.Configuration{},
			errors.New("APP_ID is required for openExchangeAPI access")
	}

	slog.Info("openExchangeAPI configured", slog.String("appID", appID))

	var cfg configuration.Configuration

	err = configuration.GetConfig("./config", &cfg)
	if err != nil {
		return configuration.Configuration{},
			fmt.Errorf("error loading configuration: %w", err)
	}

	slog.Info(cfg.Pretty())

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	return cfg, nil
}
