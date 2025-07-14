package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"main/internal/errs"
	"main/internal/errs/currency"
	"main/internal/repository"
	"main/internal/repository/memory"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	_ = iota
	beer
	gate
)

type MockWrongCurrencyRateRepo struct {
	Storage []memory.CurrencyDetails
}

func NewMockWrongStorageCurrencyRateRepo() *MockWrongCurrencyRateRepo {
	return &MockWrongCurrencyRateRepo{
		Storage: []memory.CurrencyDetails{
			beer: {0, 0},
			gate: {0, 0},
		},
	}
}

func (m *MockWrongCurrencyRateRepo) Get(currency string) (memory.CurrencyDetails, error) {
	switch currency {
	case "BEER":
		return m.Storage[beer], nil
	case "GATE":
		return m.Storage[gate], nil
	default:
		return memory.CurrencyDetails{}, errs.ErrRepoCurrencyNotFound
	}
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		currencyRateRepo repository.CurrencyRate
		errorHandler     errs.ErrorHandler
		url              string
		wantStatus       int
		wantErr          string
		wantBody         []byte
		decimalPrecision int
	}{
		{
			name:             "Test Exchange GATE to FLOKI",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=GATE&to=FLOKI&amount=123.12345",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"GATE","to":"FLOKI","amount":5923376.060924369747894400}`),
			decimalPrecision: 18,
		},
		{
			name:             "Test Exchange USDT to WBTC",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=USDT&to=WBTC&amount=1",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"USDT","to":"WBTC","amount":0.00001751}`),
			decimalPrecision: 8,
		},
		{
			name:             "Test Exchange USDT to BEER",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=USDT&to=BEER&amount=1.0",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"USDT","to":"BEER","amount":40593.254774481917919500}`),
			decimalPrecision: 18,
		},
		{
			name:             "Test Exchange BEER to USDT",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=108.108",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"BEER","to":"USDT","amount":0.002663}`),
			decimalPrecision: 6,
		},
		{
			name:             "Test Exchange FLOKI to GATE",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=FLOKI&to=GATE&amount=50",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"FLOKI","to":"GATE","amount":0.001039301310045000}`),
			decimalPrecision: 18,
		},
		{
			name:             "Test Exchange Error negative 'amount'",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=-100",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error 'amount' is not a number",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=abrakadabra",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error unknown 'from' currency MATIC",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=MATIC&to=USDT&amount=12",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error unknown 'to' currency BBB",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=BBB&amount=12",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'amount' param",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'from' param",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=&to=USDT&amount=1029",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'to' param",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=FLOKI&to=&amount=1029",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error only two params given",
			currencyRateRepo: memory.NewCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test error wrong data from currency repo",
			currencyRateRepo: NewMockWrongStorageCurrencyRateRepo(),
			errorHandler:     currency.NewErrorHandler(),
			url:              "/exchange?from=BEER&to=GATE&amount=102",
			wantStatus:       http.StatusUnprocessableEntity,
			wantErr:          "error got zero value from API or Repository",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)

			c.Request = httptest.NewRequestWithContext(
				context.Background(), "GET", tt.url, nil)

			handler := NewHandler(tt.currencyRateRepo, tt.errorHandler)
			handler.Handle(c)

			if recorder.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %d want %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantErr != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("invalid error response: %v", err)
				}

				if response["error"] != tt.wantErr {
					t.Errorf("handler returned unexpected error: got %q want %q", response["error"], tt.wantErr)
				}
			} else if tt.wantBody != nil {
				gotBody := recorder.Body.Bytes()
				if !reflect.DeepEqual(recorder.Body.Bytes(), tt.wantBody) {
					t.Errorf("error: gotBody = %s, wantBody %s", gotBody, tt.wantBody)
				}

				gotDecimalPrecision, err := getDecimalPrecision(gotBody)
				if err != nil {
					t.Errorf("error parsing decimalPrecision: %v", err)
				}

				if gotDecimalPrecision != tt.decimalPrecision {
					t.Errorf("wrond decimal precision, got = %d, want %d", gotDecimalPrecision, tt.decimalPrecision)
				}
			}
		})
	}
}

func getDecimalPrecision(body []byte) (int, error) {
	gotBodyStr := string(body)

	split := strings.Split(gotBodyStr, ".")
	if len(split) != 2 {
		return 0, errors.New("could not find float in response")
	}

	decimalPlaces := strings.Split(split[1], "}")
	if len(decimalPlaces) != 2 {
		return 0, errors.New("invalid response format")
	}

	return len(decimalPlaces[0]), nil
}
