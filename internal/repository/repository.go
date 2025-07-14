package repository

import "main/internal/repository/memory"

type CurrencyRate interface {
	Get(currency string) (memory.CurrencyDetails, error)
}
