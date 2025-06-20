package main

import (
	"context"
	"fmt"
	"log"
	"main/internal/api/openexchange"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	appID := os.Getenv("APP_ID")
	if appID == "" {
		log.Fatal(
			"personal APP_ID for openExchangeAPI is not set." +
				"Please set APP_ID env. Details in the README.md",
		)
	}

	log.Println("using appID for openExchangeAPI:", appID)

	openExchange := openexchange.New("https://openexchangerates.org/api/", appID)

	resp, err := openExchange.GetCurrencyRates(context.Background(), []string{"USD", "INR", "EUR", "BTC"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("API response: %+v\n", resp)
}
