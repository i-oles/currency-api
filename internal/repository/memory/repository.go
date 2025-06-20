package memory

import (
	"main/internal/errs"
)

type CurrencyRateRepo struct {
	Storage map[string][]float64
}

func NewCurrencyRateRepo() *CurrencyRateRepo {
	return &CurrencyRateRepo{
		Storage: map[string][]float64{
			"BEER":  {18, 0.00002461},
			"FLOKI": {18, 0.0001428},
			"GATE":  {18, 6.87},
			"USDT":  {6, 0.999},
			"WBTC":  {8, 57037.22},
		},
	}
}

func (repo *CurrencyRateRepo) Get(currency string) ([]float64, error) {
	value, ok := repo.Storage[currency]
	if !ok {
		return []float64{}, errs.ErrRepoCurrencyNotFound
	}

	return value, nil
}
