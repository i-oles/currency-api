package memory

import (
	"main/internal/errs"
)

const (
	_ = iota
	beer
	floki
	gate
	usdt
	wbtc
)

type CurrencyDetails struct {
	DecimalPrecision int
	Rate             float64
}

type CurrencyRateRepo struct {
	Storage []CurrencyDetails
}

func NewCurrencyRateRepo() *CurrencyRateRepo {
	return &CurrencyRateRepo{
		Storage: []CurrencyDetails{
			beer:  {18, 0.00002461},
			floki: {18, 0.0001428},
			gate:  {18, 6.87},
			usdt:  {6, 0.999},
			wbtc:  {8, 57037.22},
		},
	}
}

func (repo *CurrencyRateRepo) Get(currency string) (CurrencyDetails, error) {
	switch currency {
	case "BEER":
		return repo.Storage[beer], nil
	case "FLOKI":
		return repo.Storage[floki], nil
	case "GATE":
		return repo.Storage[gate], nil
	case "USDT":
		return repo.Storage[usdt], nil
	case "WBTC":
		return repo.Storage[wbtc], nil
	default:
		return CurrencyDetails{}, errs.ErrRepoCurrencyNotFound
	}
}
