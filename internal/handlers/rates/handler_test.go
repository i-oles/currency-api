package rates

import (
	"context"
	"encoding/json"
	"errors"
	"main/internal/api"
	"main/internal/errs"
	"main/internal/errs/currency"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type MockCurrencyAPI struct {
	err error
}

func NewMockAPISuccess() MockCurrencyAPI {
	return MockCurrencyAPI{}
}

func NewMockAPICurrencyNotFound() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errs.ErrCurrencyNotFound,
	}
}

func NewMockAPIInternalErr() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errors.New("random error"),
	}
}

func NewMockAPIFailureResp() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errs.ErrAPIResponse,
	}
}

func NewMockZeroValueErr() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errs.ErrZeroValue,
	}
}

func (m MockCurrencyAPI) GetCurrencyRates(
	_ context.Context, currencies []string,
) (api.Response, error) {
	if m.err != nil {
		return api.Response{}, m.err
	}

	rates := map[string]float64{
		"BDT": 122.251634,
		"BHD": 0.377252,
		"EUR": 0.869136,
		"GBP": 0.743653,
		"INR": 86.466554,
		"IRR": 42125,
		"USD": 1,
		"MRU": 0.0,
	}

	for _, currency := range currencies {
		_, ok := rates[currency]
		if !ok {
			return api.Response{}, m.err
		}
	}

	return api.Response{
		Base:      "USD",
		Rates:     rates,
		Timestamp: 1750240800,
	}, nil
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name            string
		currencyRateAPI api.CurrencyRate
		errorHandler    errs.ErrorHandler
		url             string
		wantStatus      int
		wantErr         string
		wantBody        []byte
	}{
		{
			name:            "param USD,GBP, status ok",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=USD,GBP",
			wantStatus:      http.StatusOK,
			wantBody: []byte(
				`[{"from":"USD","to":"GBP","rate":0.74365300},{"from":"GBP","to":"USD","rate":1.34471319}]`,
			),
		},
		{
			name:            "calculate for USD, GBP, EUR, status ok",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=USD,GBP,EUR",
			wantStatus:      http.StatusOK,
			wantBody: []byte(
				`[{"from":"USD","to":"GBP","rate":0.74365300},{"from":"USD","to":"EUR","rate":0.86913600},{"from":"GBP","to":"USD","rate":1.34471319},{"from":"GBP","to":"EUR","rate":1.16873865},{"from":"EUR","to":"USD","rate":1.15056792},{"from":"EUR","to":"GBP","rate":0.85562329}]`,
			),
		},
		{
			name:            "calculate for USD, BDT, BHD, INR",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=USD,BDT,BHD,INR",
			wantStatus:      http.StatusOK,
			wantBody: []byte(
				`[{"from":"USD","to":"BDT","rate":122.25163400},{"from":"USD","to":"BHD","rate":0.37725200},{"from":"USD","to":"INR","rate":86.46655400},{"from":"BDT","to":"USD","rate":0.00817985},{"from":"BDT","to":"BHD","rate":0.00308586},{"from":"BDT","to":"INR","rate":0.70728342},{"from":"BHD","to":"USD","rate":2.65074804},{"from":"BHD","to":"BDT","rate":324.05827935},{"from":"BHD","to":"INR","rate":229.20104864},{"from":"INR","to":"USD","rate":0.01156517},{"from":"INR","to":"BDT","rate":1.41386037},{"from":"INR","to":"BHD","rate":0.00436298}]`),
		},
		{
			name:         "test param USD, status 400",
			errorHandler: currency.NewErrorHandler(),
			url:          "/rates?currencies=USD",
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "test param '', status 400",
			errorHandler: currency.NewErrorHandler(),
			url:          "/rates?currencies=",
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "test param 'USD,USD', status 400",
			errorHandler: currency.NewErrorHandler(),
			url:          "/rates?currencies=",
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "test param 'USD,INR,USD', status 400",
			errorHandler: currency.NewErrorHandler(),
			url:          "/rates?currencies=",
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:            "test api failure, status 400",
			currencyRateAPI: NewMockAPIFailureResp(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=BTC,USD",
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "calculate for non existing currencies",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=AAA,BBB",
			wantStatus:      http.StatusNotFound,
			wantErr:         "error unknown currency",
		},
		{
			name:            "not all currencies exists in rates",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=GBP,AWG",
			wantStatus:      http.StatusNotFound,
			wantErr:         "error unknown currency",
		},
		{
			name:            "internal error from api module",
			currencyRateAPI: NewMockAPIInternalErr(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=GBP,BTC",
			wantStatus:      http.StatusInternalServerError,
			wantErr:         "random error",
		},
		{
			name:            "not found error from api module",
			currencyRateAPI: NewMockAPICurrencyNotFound(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=GBP,AAA",
			wantStatus:      http.StatusNotFound,
			wantErr:         "error unknown currency",
		},
		{
			name:            "error divide by zero",
			currencyRateAPI: NewMockZeroValueErr(),
			errorHandler:    currency.NewErrorHandler(),
			url:             "/rates?currencies=INR,MRU",
			wantStatus:      http.StatusUnprocessableEntity,
			wantErr:         "error got zero value from API or Repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)

			c.Request = httptest.NewRequestWithContext(
				context.Background(), "GET", tt.url, nil)

			handler := NewHandler(tt.currencyRateAPI, tt.errorHandler)
			handler.Handle(c)

			if recorder.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %d want %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantErr != "" {
				var response map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("invalid error response: %v", err)
				}

				if !strings.Contains(response["error"], tt.wantErr) {
					t.Errorf("handler returned unexpected error: got %q want %q", response["error"], tt.wantErr)
				}
			} else if tt.wantBody != nil {
				gotBody := recorder.Body.Bytes()
				if !reflect.DeepEqual(recorder.Body.Bytes(), tt.wantBody) {
					t.Errorf("calculateCurrencyRates() got = %s, want %s", gotBody, tt.wantBody)
				}
			}
		})
	}
}
