package main

import (
	"context"
	"fmt"
	"main/internal/api/openexchange"
)

func main() {
	openExchange := openexchange.New("https://openexchangerates.org/api/")

	resp, err := openExchange.GetCurrencyRates(context.Background(), []string{"USD"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("API response: %+v\n", resp)
}
