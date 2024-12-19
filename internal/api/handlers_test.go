package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"payment-gateway/internal/middleware"
	"payment-gateway/internal/models"
	"testing"
)

// ---------------- Mock Setup ------------------------------//
type mockPaymentService struct {
	shouldFail bool
}

func (m *mockPaymentService) Deposit(req *models.TransactionRequest) (*models.PaymentResult, error) {
	if m.shouldFail {
		return nil, errors.New("deposit failed")
	}
	return &models.PaymentResult{TransactionId: 1}, nil
}

func (m *mockPaymentService) Withdraw(req *models.TransactionRequest) (*models.PaymentResult, error) {
	if m.shouldFail {
		return nil, errors.New("withdrawal failed")
	}
	return &models.PaymentResult{TransactionId: 1}, nil
}

func (m *mockPaymentService) HandleCallback(callback *models.PaymentCallback) error {
	if m.shouldFail {
		return errors.New("callback failed")
	}
	return nil
}

// --------------------------------//

// Test helper functions
func setupTestHandler() (*PaymentHandler, *mockPaymentService) {
	mockService := &mockPaymentService{}

	handler := &PaymentHandler{
		paymentService: mockService,
	}
	return handler, mockService
}

func createTestRequest(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "test-key")

	// Add user context that would normally be set by auth middleware
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, 1)
	return req.WithContext(ctx)
}

func TestDeposit_ValidPayload(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDeposit_MissingAmount(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid amount" {
		t.Errorf("Expected error message 'invalid amount', got: %s", response.Message)
	}
}

func TestDeposit_InvalidAmount(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    -100,
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid amount" {
		t.Errorf("Expected error message 'invalid amount', got: %s", response.Message)
	}
}

func TestDeposit_InvalidCurrency(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "INVALID", // Invalid currency code
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid currency code" {
		t.Errorf("Expected error message 'invalid currency code', got: %s", response.Message)
	}
}

func TestDeposit_MissingCurrency(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    100.50,
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid currency code" {
		t.Errorf("Expected error message 'invalid currency code', got: %s", response.Message)
	}
}

func TestDeposit_MissingGatewayID(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid gateway id" {
		t.Errorf("Expected error message 'invalid gateway id', got: %s", response.Message)
	}
}

func TestDeposit_MissingCountryID(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/deposit", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		GatewayID: 112,
	})
	rr := httptest.NewRecorder()

	handler.Deposit(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid country id" {
		t.Errorf("Expected error message 'invalid country id', got: %s", response.Message)
	}
}

func TestWithdraw_ValidPayload(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestWithdraw_MissingAmount(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid amount" {
		t.Errorf("Expected error message 'invalid amount', got: %s", response.Message)
	}
}

func TestWithdraw_InvalidAmount(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    -100,
		Currency:  "USD",
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid amount" {
		t.Errorf("Expected error message 'invalid amount', got: %s", response.Message)
	}
}

func TestWithdraw_InvalidCurrency(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "INVALID", // Invalid currency code
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid currency code" {
		t.Errorf("Expected error message 'invalid currency code', got: %s", response.Message)
	}
}

func TestWithdraw_MissingCurrency(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    100.50,
		GatewayID: 112,
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid currency code" {
		t.Errorf("Expected error message 'invalid currency code', got: %s", response.Message)
	}
}

func TestWithdraw_MissingGatewayID(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		CountryID: 840,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid gateway id" {
		t.Errorf("Expected error message 'invalid gateway id', got: %s", response.Message)
	}
}

func TestWithdraw_MissingCountryID(t *testing.T) {
	handler, _ := setupTestHandler()

	req := createTestRequest(http.MethodPost, "/withdraw", &models.TransactionRequest{
		Amount:    100.50,
		Currency:  "USD",
		GatewayID: 112,
	})
	rr := httptest.NewRecorder()

	handler.WithdrawalHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.Message != "invalid country id" {
		t.Errorf("Expected error message 'invalid country id', got: %s", response.Message)
	}
}
