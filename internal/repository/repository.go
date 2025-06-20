package repository

type CurrencyRate interface {
	Get(currency string) ([]float64, error)
}
