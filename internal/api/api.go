package api

import "context"

type CurrencyRate interface {
	GetCurrencyRates(ctx context.Context) (Response, error)
}

type Response struct {
	Rates     map[string]float64 `json:"rates"`
	Base      string             `json:"base"`
	Timestamp int                `json:"timestamp"`
}
