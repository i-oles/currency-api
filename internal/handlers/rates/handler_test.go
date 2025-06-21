package rates

import (
	"context"
	"encoding/json"
	"errors"
	"main/internal/api"
	"main/internal/errs"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_getAllCombinations(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    [][]string
		wantErr bool
	}{
		{
			name:  "combinations for USD, GBP",
			input: []string{"USD", "GBP"},
			want: [][]string{
				{"USD", "GBP"},
				{"GBP", "USD"},
			},
		},
		{
			name:  "combinations for USD, GBP, EUR",
			input: []string{"USD", "GBP", "EUR"},
			want: [][]string{
				{"USD", "GBP"},
				{"USD", "EUR"},
				{"GBP", "USD"},
				{"GBP", "EUR"},
				{"EUR", "USD"},
				{"EUR", "GBP"},
			},
		},
		{
			name:  "combinations for USD, MMK, BTC, SLL",
			input: []string{"USD", "MMK", "BTC", "SLL"},
			want: [][]string{
				{"USD", "MMK"},
				{"USD", "BTC"},
				{"USD", "SLL"},
				{"MMK", "USD"},
				{"MMK", "BTC"},
				{"MMK", "SLL"},
				{"BTC", "USD"},
				{"BTC", "MMK"},
				{"BTC", "SLL"},
				{"SLL", "USD"},
				{"SLL", "MMK"},
				{"SLL", "BTC"},
			},
		},
		{
			name:    "combinations for one currency",
			input:   []string{"USD"},
			wantErr: true,
		},
		{
			name:    "combinations for empty list",
			input:   []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAllCombinations(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllCombinations() error = %v, wantErr %v ", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllCombinations() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockErrorHandler struct{}

func NewMockErrorHandler() *MockErrorHandler {
	return &MockErrorHandler{}
}

func (m *MockErrorHandler) Handle(c *gin.Context, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		m.sendErrorResponse(c, http.StatusGatewayTimeout, "currency rate API timeout")
	case errors.Is(err, errs.ErrAPIResponse):
		m.sendErrorResponse(c, http.StatusBadRequest, "")
	case errors.Is(err, errs.ErrCurrencyNotFound):
		m.sendErrorResponse(c, http.StatusNotFound, errs.ErrCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrRepoCurrencyNotFound):
		m.sendErrorResponse(c, http.StatusNotFound, errs.ErrRepoCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrBadRequest):
		m.sendErrorResponse(c, http.StatusBadRequest, "")
	default:
		m.sendErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
}

func (m *MockErrorHandler) sendErrorResponse(c *gin.Context, status int, message string) {
	if message == "" {
		c.AbortWithStatus(status)
	} else {
		c.JSON(status, gin.H{"error": message})
	}
}

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
		wantBody        []map[string]interface{}
	}{
		{
			name:            "test param USD,GBP, status ok",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=USD,GBP",
			wantStatus:      http.StatusOK,
			wantBody: []map[string]interface{}{
				{
					"from": "GBP", "to": "USD", "rate": 1.344713,
				},
				{
					"from": "USD", "to": "GBP", "rate": 0.743653,
				},
			},
			wantErr: "",
		},
		{
			name:            "calculate for USD, GBP, EUR, status ok",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=USD,GBP,EUR",
			wantStatus:      http.StatusOK,
			wantBody: []map[string]interface{}{
				{
					"from": "USD", "to": "EUR", "rate": 0.869136,
				},
				{
					"from": "USD", "to": "GBP", "rate": 0.743653,
				},
				{
					"from": "GBP", "to": "USD", "rate": 1.344713,
				},
				{
					"from": "GBP", "to": "EUR", "rate": 1.168739,
				},
				{
					"from": "EUR", "to": "USD", "rate": 1.150568,
				},
				{
					"from": "EUR", "to": "GBP", "rate": 0.855623,
				},
			},
		},
		{
			name:            "calculate for USD, BDT, BHD, INR",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=USD,BDT,BHD,INR",
			wantStatus:      http.StatusOK,
			wantBody: []map[string]interface{}{
				{
					"from": "USD", "to": "BDT", "rate": 122.251634,
				},
				{
					"from": "USD", "to": "BHD", "rate": 0.377252,
				},
				{
					"from": "USD", "to": "INR", "rate": 86.466554,
				},
				{
					"from": "BHD", "to": "USD", "rate": 2.650748,
				},
				{
					"from": "BHD", "to": "INR", "rate": 229.201049,
				},
				{
					"from": "BHD", "to": "BDT", "rate": 324.058279,
				},
				{
					"from": "BDT", "to": "BHD", "rate": 0.003086,
				},
				{
					"from": "BDT", "to": "INR", "rate": 0.707283,
				},
				{
					"from": "BDT", "to": "USD", "rate": 0.00818,
				},
				{
					"from": "INR", "to": "BDT", "rate": 1.41386,
				},
				{
					"from": "INR", "to": "BHD", "rate": 0.004363,
				},
				{
					"from": "INR", "to": "USD", "rate": 0.011565,
				},
			},
		},
		{
			name:         "test param USD, status 400",
			url:          "/rates?currencies=USD",
			errorHandler: NewMockErrorHandler(),
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:         "test param '', status 400",
			url:          "/rates?currencies=",
			errorHandler: NewMockErrorHandler(),
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:            "test api failure, status 400",
			currencyRateAPI: NewMockAPIFailureResp(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=BTC,USD",
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "calculate for non existing currencies",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=AAA,BBB",
			wantStatus:      http.StatusNotFound,
			wantBody:        []map[string]interface{}{},
			wantErr:         "error unknown currency",
		},
		{
			name:            "not all currencies exists in rates",
			currencyRateAPI: NewMockAPISuccess(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=GBP,AWG",
			wantStatus:      http.StatusNotFound,
			wantBody:        []map[string]interface{}{},
			wantErr:         "error unknown currency",
		},
		{
			name:            "internal error from api module",
			currencyRateAPI: NewMockAPIInternalErr(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=GBP,BTC",
			wantStatus:      http.StatusInternalServerError,
			wantBody:        []map[string]interface{}{},
			wantErr:         "random error",
		},
		{
			name:            "not found error from api module",
			currencyRateAPI: NewMockAPICurrencyNotFound(),
			errorHandler:    NewMockErrorHandler(),
			url:             "/rates?currencies=GBP,AAA",
			wantStatus:      http.StatusNotFound,
			wantBody:        []map[string]interface{}{},
			wantErr:         "error unknown currency",
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

				if response["error"] != tt.wantErr {
					t.Errorf("handler returned unexpected error: got %q want %q", response["error"], tt.wantErr)
				}
			} else if tt.wantBody != nil {
				var gotBody []map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &gotBody); err != nil {
					t.Fatalf("handler returned wrong body: %v", err)
				}

				sortSlice(gotBody)
				sortSlice(tt.wantBody)

				if !slicesEqual(gotBody, tt.wantBody) {
					t.Errorf("calculateCurrencyRates() got = %v, want %v", gotBody, tt.wantBody)
				}
			}
		})
	}
}

const epsilon = 1e-6

func mapsEqual(got, want map[string]interface{}) bool {
	if got["from"] != want["from"] || got["to"] != want["to"] {
		return false
	}

	gotAny, ok := got["rate"]
	if !ok {
		return false
	}

	wantAny, ok := want["rate"]
	if !ok {
		return false
	}

	gotRate, okGot := gotAny.(float64)
	wantRate, okWant := wantAny.(float64)

	if !okGot || !okWant {
		return false
	}

	return math.Abs(gotRate-wantRate) < epsilon
}

func slicesEqual(got, want []map[string]interface{}) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if !mapsEqual(got[i], want[i]) {
			return false
		}
	}

	return true
}

func sortSlice(slice []map[string]interface{}) {
	sort.Slice(slice, func(i, j int) bool {
		mapI, mapJ := slice[i], slice[j]

		if mapI == nil || mapJ == nil {
			return false
		}

		valI, okI := mapI["rate"]
		if !okI {
			return false
		}

		valJ, okJ := mapJ["rate"]
		if !okJ {
			return false
		}

		valIFloat, okI := valI.(float64)
		valJFloat, okJ := valJ.(float64)

		if !okI || !okJ {
			return false
		}

		return valIFloat < valJFloat
	})
}

func Test_calculateCurrencyRates(t *testing.T) {
	type args struct {
		rates                map[string]float64
		currencyCombinations [][]string
	}

	tests := []struct {
		name    string
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "not enough currencies in combination",
			args: args{
				currencyCombinations: [][]string{},
			},
			wantErr: true,
		},
		{
			name: "too many currencies in combination",
			args: args{
				currencyCombinations: [][]string{
					[]string{"GBP", "USD", "BTC"},
				},
			},
			wantErr: true,
		},
		{
			name: "not enough currencies in combination",
			args: args{
				currencyCombinations: [][]string{
					[]string{"BTC"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateCurrencyRates(tt.args.rates, tt.args.currencyCombinations)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateCurrencyRates() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateCurrencyRates() got = %v, want %v", got, tt.want)
			}
		})
	}
}
