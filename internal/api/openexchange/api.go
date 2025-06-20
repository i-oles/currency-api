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

type OpenExchange struct {
	apiURL        string
	apiSourceFile string
	baseCurrency  string
}

func New(apiURL string) OpenExchange {
	return OpenExchange{
		apiURL:        apiURL,
		apiSourceFile: "latest.json",
		baseCurrency:  "USD",
	}
}

func (o OpenExchange) GetCurrencyRates(
	ctx context.Context,
	currencies []string,
) (api.Response, error) {
	reqURL, err := url.Parse(o.apiURL)
	if err != nil {
		return api.Response{}, fmt.Errorf("error parsing api url %s: %w", o.apiURL, err)
	}

	reqURL.Path = path.Join(reqURL.Path, o.apiSourceFile)

	query := url.Values{}
	query.Set("app_id", AppID)
	query.Set("base", o.baseCurrency)

	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return api.Response{}, fmt.Errorf("error creating request %s: %w", reqURL.String(), err)
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
