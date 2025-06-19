package rates

import (
	"context"
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
			name: "calculate for USD, GBP, EUR",
			args: args{
				rates: map[string]float64{
					"USD": 1.0,
					"GBP": 0.743653,
					"EUR": 0.869136,
				},
				currencyCombinations: [][]string{
					{"USD", "GBP"},
					{"USD", "EUR"},
					{"EUR", "GBP"},
					{"EUR", "USD"},
					{"GBP", "USD"},
					{"GBP", "EUR"},
				},
			},
			want: []map[string]interface{}{
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
			name: "calculate for USD, GBP",
			args: args{
				rates: map[string]float64{
					"USD": 1.0,
					"GBP": 0.7434,
				},
				currencyCombinations: [][]string{
					{"USD", "GBP"},
					{"GBP", "USD"},
				},
			},
			want: []map[string]interface{}{
				{
					"from": "USD", "to": "GBP", "rate": 0.743400,
				},
				{
					"from": "GBP", "to": "USD", "rate": 1.345171,
				},
			},
		},
		{
			name: "calculate for GBP, EUR",
			args: args{
				rates: map[string]float64{
					"GBP": 0.743653,
					"EUR": 0.869136,
				},
				currencyCombinations: [][]string{
					{"EUR", "GBP"},
					{"GBP", "EUR"},
				},
			},
			want: []map[string]interface{}{
				{
					"from": "GBP", "to": "EUR", "rate": 1.168739,
				},
				{
					"from": "EUR", "to": "GBP", "rate": 0.855623,
				},
			},
		},
		{
			name: "calculate for GBP",
			args: args{
				rates: map[string]float64{
					"GBP": 0.743653,
				},
				currencyCombinations: [][]string{
					{"GBP"},
				},
			},
			wantErr: true,
		},
		{
			name: "calculate for empty rates",
			args: args{
				rates: map[string]float64{},
				currencyCombinations: [][]string{
					{"BTC", "INR"},
					{"INR", "BTC"},
				},
			},
			wantErr: true,
		},
		{
			name: "calculate for empty rates, and empty combinations",
			args: args{
				rates:                nil,
				currencyCombinations: [][]string{},
			},
			wantErr: true,
		},
		{
			name: "not all currencies exists in rates",
			args: args{
				rates: map[string]float64{
					"GBP": 0.743653,
				},
				currencyCombinations: [][]string{
					{"GBP", "INR"},
					{"INR", "GBP"},
				},
			},
			wantErr: true,
		},
		{
			name: "calculate for non existing rates",
			args: args{
				rates: map[string]float64{
					"GBP": 0.743653,
					"EUR": 0.869136,
				},
				currencyCombinations: [][]string{
					{
						"AAA", "BBB",
						"BBB", "AAA",
					},
				},
			},
			wantErr: true,
		},

		{
			name: "calculate for USD, BDT, BHC, INR",
			args: args{
				rates: map[string]float64{
					"USD": 1.0,
					"BDT": 122.251634,
					"BHD": 0.377252,
					"INR": 86.466554,
				},
				currencyCombinations: [][]string{
					{"USD", "BDT"},
					{"USD", "BHD"},
					{"USD", "INR"},
					{"BHD", "USD"},
					{"BHD", "INR"},
					{"BHD", "BDT"},
					{"BDT", "BHD"},
					{"BDT", "INR"},
					{"BDT", "USD"},
					{"INR", "USD"},
					{"INR", "BHD"},
					{"INR", "BDT"},
				},
			},

			want: []map[string]interface{}{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateCurrencyRates(tt.args.rates, tt.args.currencyCombinations)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateCurrencyRates() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			sortSlice(got)
			sortSlice(tt.want)

			if !slicesEqual(got, tt.want) {
				t.Errorf("calculateCurrencyRates() got = %v, want %v", got, tt.want)
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

//TODO remove me?
//func Test_validateParameter(t *testing.T) {
//	tests := []struct {
//		name    string
//		param   string
//		wantErr bool
//	}{
//		{
//			name:  "validation ok for USD,GBP",
//			param: "USD,GBP",
//		},
//		{
//			name:  "validation ok for USD,GBP,PLN,EUR,INR",
//			param: "USD,GBP,PLN,EUR,INR",
//		},
//		{
//			name:    "error validation for USD",
//			param:   "USD",
//			wantErr: true,
//		},
//		{
//			name:    "error validation for empty string",
//			param:   "",
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := validateParameter(tt.param); (err != nil) != tt.wantErr {
//				t.Errorf("validateParameter() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

type MockCurrencyAPI struct{}

func NewMockCurrencyAPI() MockCurrencyAPI {
	return MockCurrencyAPI{}
}

func (m MockCurrencyAPI) GetCurrencyRates(
	_ context.Context, _ []string,
) (api.Response, error) {
	return api.Response{
		Base: "USD",
		Rates: map[string]float64{
			"DKK": 6.48352,
			"DOP": 59.202312,
			"DZD": 130.22232,
			"EGP": 50.458337,
			"ERN": 15,
			"ETB": 134.8,
			"EUR": 0.869289,
			"FJD": 2.24725,
			"FKP": 0.742708,
			"GBP": 0.742708,
			"GEL": 2.72,
			"GGP": 0.742708,
			"GHS": 10.295649,
			"GIP": 0.742708,
			"GMD": 71.500005,
			"GNF": 8658.789126,
			"GTQ": 7.677452,
			"GYD": 209.058301,
			"IMP": 0.742708,
			"INR": 86.466554,
			"IQD": 1309.719481,
			"IRR": 42125,
			"UAH": 41.534469,
			"UGX": 3593.775173,
			"USD": 1,
			"UYU": 40.984695,
		},
		Timestamp: 1750240800,
	}, nil
}

type MockFailureCurrencyAPI struct{}

func NewMockFailureCurrencyAPI() MockFailureCurrencyAPI {
	return MockFailureCurrencyAPI{}
}

func (m MockFailureCurrencyAPI) GetCurrencyRates(
	_ context.Context,
	_ []string,
) (api.Response, error) {
	return api.Response{}, errs.APIResponseError("error from API", errors.New("failure"))
}

func TestHandler_Handle(t *testing.T) {
	type fields struct {
		currencyRateAPI api.CurrencyRate
	}

	tests := []struct {
		name        string
		fields      fields
		queryParams string
		wantStatus  int
	}{
		{
			name: "test param USD,GBP, status ok",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			queryParams: "USD,GBP",
			wantStatus:  http.StatusOK,
		},
		{
			name: "test param USD, status 400",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			queryParams: "USD",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "test param '', status 400",
			fields: fields{
				currencyRateAPI: NewMockCurrencyAPI(),
			},
			queryParams: "",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "test api failure, status 400",
			fields: fields{
				currencyRateAPI: NewMockFailureCurrencyAPI(),
			},
			queryParams: "BTC,USD",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)

			c.Request = httptest.NewRequestWithContext(
				context.Background(), "GET", "/rates?currencies="+tt.queryParams, nil)

			handler := NewHandler(tt.fields.currencyRateAPI)
			handler.Handle(c)

			if recorder.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %d want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
