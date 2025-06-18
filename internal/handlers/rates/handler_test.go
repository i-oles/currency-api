package rates

import (
	"reflect"
	"sort"
	"testing"
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
			name: "calculate for USD, GBP, EUR, SLL",
			args: args{
				rates: map[string]float64{
					"USD": 1.0,
					"MMK": 2098,
					"BTC": 0.000009548793,
					"SLL": 20969.5,
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

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateCurrencyRates() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func sortSlice(slice []map[string]interface{}) {
	sort.Slice(slice, func(i, j int) bool {
		valI := slice[i]["rate"].(float64)
		valJ := slice[j]["rate"].(float64)

		return valI < valJ
	})
}

func Test_getAllCombinations(t *testing.T) {
	type args struct {
		input []string
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "combinations for USD, GBP",
			args: args{
				input: []string{"USD", "GBP"},
			},
			want: [][]string{
				{"USD", "GBP"},
				{"GBP", "USD"},
			},
		},
		{
			name: "combinations for USD, GBP, EUR",
			args: args{
				input: []string{"USD", "GBP", "EUR"},
			},
			want: [][]string{
				{"USD", "GBP"},
				{"USD", "EUR"},
				{"EUR", "GBP"},
				{"EUR", "USD"},
				{"GBP", "USD"},
				{"GBP", "EUR"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAllCombinations(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllCombinations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateParameter(t *testing.T) {
	type args struct {
		currencies []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validation ok for USD, GBP",
			args: args{
				currencies: []string{"USD", "GBP"},
			},
		},
		{
			name: "validation ok for USD, GBP, PLN, EUR, INR",
			args: args{
				currencies: []string{"USD", "GBP", "PLN", "EUR", "INR"},
			},
		},
		{
			name: "error validation for 'USD'",
			args: args{
				currencies: []string{"USD"},
			},
			wantErr: true,
		},
		{
			name: "error validation for empty string",
			args: args{
				currencies: []string{""},
			},
			wantErr: true,
		},
		{
			name: "error validation for empty parameter",
			args: args{
				currencies: []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateParameter(tt.args.currencies); (err != nil) != tt.wantErr {
				t.Errorf("validateParameter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
