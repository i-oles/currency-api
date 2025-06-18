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
				{"GBP", "USD"},
				{"GBP", "EUR"},
				{"EUR", "USD"},
				{"EUR", "GBP"},
			},
		},
		{
			name: "combinations for USD, MMK, BTC, SLL",
			args: args{
				input: []string{"USD", "MMK", "BTC", "SLL"},
			},
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
