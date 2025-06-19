package rates

import (
	"context"
	"encoding/json"
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

type MockCurrencyAPI struct {
	err error
}

func NewMockCurrencyAPI() MockCurrencyAPI {
	return MockCurrencyAPI{}
}
func NewMockCurrencyNotFoundErrAPI() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errs.ErrCurrencyNotFound,
	}
}

func NewMockApiRespErrCurrencyAPI() MockCurrencyAPI {
	return MockCurrencyAPI{
		err: errs.ErrAPIResponse,
	}
}

func NewMockNotEnoughCurrenciesAPI() MockCurrencyAPI {
	return MockCurrencyAPI{}
}

func NewMockEmptyParamCurrencyAPI() MockCurrencyAPI {
	return MockCurrencyAPI{}
}

func (m MockCurrencyAPI) GetCurrencyRates(
	_ context.Context, _ []string,
) (api.Response, error) {
	if m.err != nil {
		return api.Response{}, m.err
	}
	return api.Response{
		Base: "USD",
		Rates: map[string]float64{
			"BDT": 122.251634,
			"BHD": 0.377252,
			"DKK": 6.48352,
			"DOP": 59.202312,
			"DZD": 130.22232,
			"EGP": 50.458337,
			"ERN": 15,
			"EUR": 0.869136,
			"ETB": 134.8,
			"FJD": 2.24725,
			"FKP": 0.742708,
			"GBP": 0.743653,
			"GEL": 2.72,
			"GGP": 0.742708,
			"GHS": 10.295649,
			"GIP": 0.742708,
			"GMD": 71.500005,
			"GNF": 8658.789126,
			"GTQ": 7.677452,
			"GYD": 209.058301,
			"IMP": 0.742708,
			"IQD": 1309.719481,
			"INR": 86.466554,
			"IRR": 42125,
			"UAH": 41.534469,
			"UGX": 3593.775173,
			"USD": 1,
			"UYU": 40.984695,
		},
		Timestamp: 1750240800,
	}, nil
}

func TestHandler_Handle(t *testing.T) {
	type fields struct {
		currencyRateAPI api.CurrencyRate
	}

	tests := []struct {
		name       string
		fields     fields
		url        string
		wantStatus int
		wantErr    string
		wantBody   []map[string]interface{}
	}{
		{
			name: "test param USD,GBP, status ok",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			url:        "/rates?currencies=USD,GBP",
			wantStatus: http.StatusOK,
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
			name: "calculate for USD, GBP, EUR, status ok",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			url:        "/rates?currencies=USD,GBP,EUR",
			wantStatus: http.StatusOK,
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
			name: "calculate for USD, BDT, BHD, INR",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			url:        "/rates?currencies=USD,BDT,BHD,INR",
			wantStatus: http.StatusOK,
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
			name: "test param USD, status 400",
			fields: fields{
				currencyRateAPI: NewMockNotEnoughCurrenciesAPI(),
			},
			url:        "/rates?currencies=USD",
			wantStatus: http.StatusBadRequest,
			wantBody:   []map[string]interface{}{},
		},
		{
			name: "test param '', status 400",
			fields: fields{
				currencyRateAPI: NewMockEmptyParamCurrencyAPI(),
			},
			url:        "/rates?currencies=",
			wantStatus: http.StatusBadRequest,
			wantBody:   []map[string]interface{}{},
		},
		{
			name: "test api failure, status 400",
			fields: fields{
				currencyRateAPI: NewMockApiRespErrCurrencyAPI(),
			},
			url:        "/rates?currencies=BTC,USD",
			wantStatus: http.StatusBadRequest,
			wantBody:   []map[string]interface{}{},
		},
		{
			name: "calculate for non existing currencies",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			url:        "/rates?currencies=AAA,BBB",
			wantStatus: http.StatusNotFound,
			wantBody:   []map[string]interface{}{},
			wantErr:    errs.ErrCurrencyNotFound.Error(),
		},
		{
			name: "not all currencies exists in rates",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			url:        "/rates?currencies=GBP,AWG",
			wantStatus: http.StatusNotFound,
			wantBody:   []map[string]interface{}{},
			wantErr:    errs.ErrCurrencyNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)

			c.Request = httptest.NewRequestWithContext(
				context.Background(), "GET", tt.url, nil)

			handler := NewHandler(tt.fields.currencyRateAPI)
			handler.Handle(c)

			if recorder.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %d want %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantBody == nil {
				var response map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("invalid error response: %v", err)
				}

				if response["error"] != tt.wantErr {
					t.Errorf("handler returned unexpected error: got %q want %q", response["error"], tt.wantErr)
				}

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

	gotRate := got["rate"].(float64)
	wantRate := want["rate"].(float64)

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
		valI := slice[i]["rate"].(float64)
		valJ := slice[j]["rate"].(float64)

		return valI < valJ
	})
}
