package openexchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/internal/api"
	"main/internal/errs"
	"net/http"
	"net/url"
	"path"
)

const (
	apiSourceFile = "latest.json"
	baseCurrency  = "USD"
)

type OpenExchange struct {
	URL *url.URL
}

func New(apiURL, apiAppID string) (OpenExchange, error) {
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return OpenExchange{}, fmt.Errorf("error parsing api url %s: %w", apiURL, err)
	}

	reqURL.Path = path.Join(reqURL.Path, apiSourceFile)

	query := url.Values{}
	query.Set("app_id", apiAppID)
	query.Set("base", baseCurrency)

	reqURL.RawQuery = query.Encode()

	return OpenExchange{
		URL: reqURL,
	}, nil
}

func (o OpenExchange) GetCurrencyRates(
	ctx context.Context,
	currencies []string,
) (api.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.URL.String(), nil)
	if err != nil {
		return api.Response{}, fmt.Errorf("error creating request %s: %w", o.URL.String(), err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return api.Response{}, errs.ErrAPIResponse
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.Response{}, fmt.Errorf("error reading response body %s: %w", resp.Body, err)
	}

	var result api.Response

	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return api.Response{}, fmt.Errorf("error unmarshaling response body %s: %w", bodyBytes, err)
	}

	neededCurrencies := make(map[string]float64, len(currencies))

	for _, currency := range currencies {
		val, ok := result.Rates[currency]
		if !ok {
			return api.Response{}, errs.ErrCurrencyNotFound
		}

		neededCurrencies[currency] = val
	}

	result.Rates = neededCurrencies

	return result, nil
}
