package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"main/internal/errs"
	"main/internal/repository"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type MockCurrencyRateRepo struct {
	Storage map[string][]float64
}

func NewMockCurrencyRateRepo() *MockCurrencyRateRepo {
	return &MockCurrencyRateRepo{
		Storage: map[string][]float64{
			"BEER":  {18, 0.00002461},
			"FLOKI": {18, 0.0001428},
			"GATE":  {18, 6.87},
			"USDT":  {6, 0.999},
			"WBTC":  {8, 57037.22},
		},
	}
}

func (m *MockCurrencyRateRepo) Get(currency string) ([]float64, error) {
	value, ok := m.Storage[currency]
	if !ok {
		return []float64{}, errs.ErrRepoCurrencyNotFound
	}

	return value, nil
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
		m.sendErrorResponse(c, errs.StatusCode400, "")
	case errors.Is(err, errs.ErrCurrencyNotFound):
		m.sendErrorResponse(c, http.StatusNotFound, errs.ErrCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrRepoCurrencyNotFound):
		m.sendErrorResponse(c, http.StatusBadRequest, errs.ErrRepoCurrencyNotFound.Error())
	case errors.Is(err, errs.ErrBadRequest):
		m.sendErrorResponse(c, http.StatusBadRequest, "")
	default:
		m.sendErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
}

func (m *MockErrorHandler) sendErrorResponse(c *gin.Context, status int, message string) {
	if message == "" {
		c.JSON(status, nil)
	} else {
		c.JSON(status, gin.H{"error": message})
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
		wantBody         map[string]interface{}
	}{
		//{
		//	name:             "Test Exchange GATE to FLOKI",
		//	currencyRateRepo: NewMockCurrencyRateRepo(),
		//	errorHandler:     NewMockErrorHandler(),
		//	url:              "/exchange?from=GATE&to=FLOKI&amount=123.12345",
		//	wantStatus:       http.StatusOK,
		//	wantBody: map[string]interface{}{
		//		"from":   "GATE",
		//		"to":     "FLOKI",
		//		"amount": 5923376.060924369747894400,
		//	},
		//},
		//{
		//	name:             "Test Exchange USDT to WBTC",
		//	currencyRateRepo: NewMockCurrencyRateRepo(),
		//	errorHandler:     NewMockErrorHandler(),
		//	url:              "/exchange?from=USDT&to=WBTC&amount=1",
		//	wantStatus:       http.StatusOK,
		//	wantBody: map[string]string{
		//		"from":   "USDT",
		//		"to":     "WBTC",
		//		"amount": "0.00001751",
		//	},
		//},
		{
			name:             "Test Exchange USDT to BEER",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=USDT&to=BEER&amount=1.0",
			wantStatus:       http.StatusOK,
			wantBody: map[string]interface{}{
				"from":   "USDT",
				"to":     "BEER",
				"amount": 40593.254774481917919500,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			fmt.Printf("tt.wantBody: %+v\n", tt.wantBody)

			c, _ := gin.CreateTestContext(recorder)

			c.Request = httptest.NewRequestWithContext(
				context.Background(), "GET", tt.url, nil)

			handler := NewHandler(tt.currencyRateRepo, tt.errorHandler)
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
				var gotBody map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &gotBody); err != nil {
					t.Fatalf("handler returned wrong body: %v", err)
				}

				fmt.Printf("gotBody: %+v\n", gotBody)

				if !mapsEqual(gotBody, tt.wantBody) {
					t.Errorf("calculateCurrencyRates() got = %v, want %v", gotBody, tt.wantBody)
				}
			}
		})
	}
}
