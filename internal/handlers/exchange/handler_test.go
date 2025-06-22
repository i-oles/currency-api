package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"main/internal/errs"
	"main/internal/repository"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func NewMockWrongStorageCurrencyRateRepo() *MockCurrencyRateRepo {
	return &MockCurrencyRateRepo{
		Storage: map[string][]float64{
			"FLOKI": {18},
			"GATE":  {6.87},
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
	case errors.Is(err, errs.ErrRepoCurrencyNotFound),
		errors.Is(err, errs.ErrNegativeAmount),
		errors.Is(err, errs.ErrAmountNotNumber),
		errors.Is(err, errs.ErrEmptyParam):
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

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name             string
		currencyRateRepo repository.CurrencyRate
		errorHandler     errs.ErrorHandler
		url              string
		wantStatus       int
		wantErr          string
		wantBody         []byte
	}{
		{
			name:             "Test Exchange GATE to FLOKI",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=GATE&to=FLOKI&amount=123.12345",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"GATE","to":"FLOKI","amount":5923376.060924369747894400}`),
		},
		{
			name:             "Test Exchange USDT to WBTC",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=USDT&to=WBTC&amount=1",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"USDT","to":"WBTC","amount":0.00001751}`),
		},
		{
			name:             "Test Exchange USDT to BEER",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=USDT&to=BEER&amount=1.0",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"USDT","to":"BEER","amount":40593.254774481917919500}`),
		},
		{
			name:             "Test Exchange BEER to USDT",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=108.108",
			wantStatus:       http.StatusOK,
			wantBody:         []byte(`{"from":"BEER","to":"USDT","amount":0.002663}`),
		},
		{
			name:             "Test Exchange Error negative 'amount'",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=-100",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error 'amount' is not a number",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=abrakadabra",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error unknown 'from' currency MATIC",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=MATIC&to=USDT&amount=12",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error unknown 'to' currency BBB",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=BBB&amount=12",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'amount' param",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT&amount=",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'from' param",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=&to=USDT&amount=1029",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error empty 'to' param",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=FLOKI&to=&amount=1029",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test Exchange Error only two params given",
			currencyRateRepo: NewMockCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=BEER&to=USDT",
			wantStatus:       http.StatusBadRequest,
			wantBody:         nil,
		},
		{
			name:             "Test wrong data from currency repo",
			currencyRateRepo: NewMockWrongStorageCurrencyRateRepo(),
			errorHandler:     NewMockErrorHandler(),
			url:              "/exchange?from=FLOKI&to=GATE&amount=102",
			wantStatus:       http.StatusInternalServerError,
			wantErr:          "len of currency details is invalid",
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
			}
		})
	}
}
