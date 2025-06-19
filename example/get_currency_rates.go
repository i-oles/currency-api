package main

import (
	"context"
	"fmt"
	"main/internal/api/openexchange"
)

func main() {
	openExchange := openexchange.New()

	resp, err := openExchange.GetCurrencyRates(context.Background(), []string{"USD", "EUR"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("API response: %+v\n", resp)
}
