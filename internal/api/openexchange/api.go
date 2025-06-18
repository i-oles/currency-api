package openexchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/internal/api"
	"net/http"
	"net/url"
	"path"
)

type OpenExchange struct {
	apiURL        string
	apiSourceFile string
	baseCurrency  string
}

func New() OpenExchange {
	return OpenExchange{
		//TODO: move url to config?
		apiURL:        "https://openexchangerates.org/api/",
		apiSourceFile: "latest.json",
		baseCurrency:  "USD",
	}
}

func (o OpenExchange) GetCurrencyRates(ctx context.Context) (api.Response, error) {
	baseURL, err := url.Parse(o.apiURL)
	if err != nil {
		return api.Response{}, fmt.Errorf("error parsing api url %s: %w", o.apiURL, err)
	}

	baseURL.Path = path.Join(baseURL.Path, o.apiSourceFile)

	query := url.Values{}
	query.Set("app_id", AppID)
	query.Set("base", o.baseCurrency)

	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return api.Response{}, fmt.Errorf("error creating request %s: %w", baseURL.String(), err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return api.Response{}, fmt.Errorf("error making request %s: %w", baseURL.String(), err)
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

	return result, nil
}
