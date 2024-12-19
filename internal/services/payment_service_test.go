package services

import (
	"context"
	"errors"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"testing"
	"time"
)

// ----------- Mock Setup ------------------//
type mockComplianceService struct {
	shouldPass bool
}

func (m *mockComplianceService) CheckStatus(req *models.TransactionRequest) (string, error) {
	if !m.shouldPass {
		return "failed", errors.New("compliance check failed")
	}
	return "passed", nil
}

// Mock AccountService
type mockAccountService struct {
	balance float64
}

func (m *mockAccountService) GetBalance(userID int) (float64, error) {
	return m.balance, nil
}

// Mock PaymentGateway
type mockPaymentGateway struct {
	shouldFail    bool
	shouldTimeout bool
}

func (m *mockPaymentGateway) ProcessPayment(ctx context.Context, trx *db.Transaction) (*GatewayResult, error) {
	if m.shouldFail && m.shouldTimeout {
		<-time.After(1 * time.Second)
		return nil, errors.New("payment gateway error.hehe")

	}
	if m.shouldFail {
		return nil, errors.New("payment processing failed")
	}

	return &GatewayResult{
		GatewayTxnId: "mock_txn_123",
	}, nil
}

type mockTransactionRepository struct {
	transactions map[string]*db.Transaction
	lastID       int
}

func newMockRepository() *mockTransactionRepository {
	return &mockTransactionRepository{
		transactions: make(map[string]*db.Transaction),
		lastID:       0,
	}
}

func (m *mockTransactionRepository) Create(tx *db.Transaction) (*db.Transaction, error) {
	m.lastID++
	tx.ID = m.lastID
	m.transactions[tx.GatewayTxnId] = tx
	return tx, nil
}

func (m *mockTransactionRepository) Update(tx db.Transaction) error {
	if _, exists := m.transactions[tx.GatewayTxnId]; !exists {
		return errors.New("transaction not found")
	}
	m.transactions[tx.GatewayTxnId] = &tx
	return nil
}

func (m *mockTransactionRepository) GetTransactionByGatewayTxnId(gatewayTxnId string) (*db.Transaction, error) {
	if tx, exists := m.transactions[gatewayTxnId]; exists {
		return tx, nil
	}
	return nil, nil
}

// ---------------------------------------------- //

// Test setup helper
func setupTestService(t *testing.T, compliancePass bool, balance float64) (*paymentService, *mockPaymentGateway, *mockTransactionRepository) {
	mockGateway := &mockPaymentGateway{shouldFail: !compliancePass}
	mockRepo := newMockRepository()

	// Override the service creation
	service := &paymentService{
		cs:   &mockComplianceService{shouldPass: compliancePass},
		as:   &mockAccountService{balance: balance},
		repo: mockRepo,
	}

	// Store original gateway function
	originalGateway := GetPaymentGateway

	// Override gateway for testing
	GetPaymentGateway = func(countryId int, gatewayId int) PaymentGateway {
		return mockGateway
	}

	// Restore after test
	t.Cleanup(func() {
		GetPaymentGateway = originalGateway
	})

	return service, mockGateway, mockRepo
}

func TestDeposit_Success(t *testing.T) {
	service, mockGateway, mockRepo := setupTestService(t, true, 1000)
	mockGateway.shouldFail = false

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	result, err := service.Deposit(req)
	if err != nil {
		t.Errorf("Expected successful deposit, got error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// Verify the saved transaction
	savedTx, err := mockRepo.GetTransactionByGatewayTxnId("mock_txn_123")
	if err != nil {
		t.Errorf("Failed to fetch transaction: %v", err)
	}
	if savedTx == nil {
		t.Fatal("Transaction was not saved")
	}

	// Verify transaction details
	if savedTx.Amount != req.Amount {
		t.Errorf("Expected amount %v, got %v", req.Amount, savedTx.Amount)
	}
	if savedTx.UserID != req.UserID {
		t.Errorf("Expected userID %v, got %v", req.UserID, savedTx.UserID)
	}
	if savedTx.Type != "deposit" {
		t.Errorf("Expected type 'deposit', got %v", savedTx.Type)
	}
	if savedTx.Status != "pending" {
		t.Errorf("Expected status 'pending', got %v", savedTx.Status)
	}
	if savedTx.GatewayID != req.GatewayID {
		t.Errorf("Expected gatewayID %v, got %v", req.GatewayID, savedTx.GatewayID)
	}
	if savedTx.GatewayTxnId != "mock_txn_123" {
		t.Errorf("Expected gatewayTxnId 'mock_txn_123', got %v", savedTx.GatewayTxnId)
	}
}

func TestDeposit_PaymentProcessingFailure(t *testing.T) {
	service, mockGateway, _ := setupTestService(t, true, 1000)
	mockGateway.shouldFail = true

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	result, err := service.Deposit(req)
	if err == nil {
		t.Error("Expected error due to payment processing failure, got success")
	}
	if result != nil {
		t.Error("Expected nil result, got result")
	}
	if err != nil && err.Error() != "Payment gateway error." {
		t.Errorf("Expected payment processing error, got: %v", err)
	}
}

func TestDeposit_ComplianceFailure(t *testing.T) {
	service, _, _ := setupTestService(t, false, 1000)

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	_, err := service.Deposit(req)
	if err == nil {
		t.Error("Expected compliance check to fail, but got success")
	}
	if err != nil && err.Error() != "compliance check failed" {
		t.Errorf("Expected compliance error, got: %v", err)
	}
}

func TestDeposit_PaymentProcessingTimeout(t *testing.T) {
	service, pg, _ := setupTestService(t, true, 1000)
	pg.shouldFail = true
	pg.shouldTimeout = true

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}
	res, err := service.Deposit(req)

	if res != nil {
		t.Errorf("Expected a nil response but received value")
	}

	if err != nil && err.Error() != "Payment gateway error." {
		t.Errorf("Expected payment gateway error, got: %v", err)
	}

}

func TestWithdraw_Success(t *testing.T) {
	service, mockGateway, mockRepo := setupTestService(t, true, 1000)
	mockGateway.shouldFail = false

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	result, err := service.Withdraw(req)
	if err != nil {
		t.Errorf("Expected successful withdrawal, got error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}

	// Verify the saved transaction
	savedTx, err := mockRepo.GetTransactionByGatewayTxnId("mock_txn_123")
	if err != nil {
		t.Errorf("Failed to fetch transaction: %v", err)
	}
	if savedTx == nil {
		t.Fatal("Transaction was not saved")
	}

	// Verify transaction details
	if savedTx.Amount != req.Amount {
		t.Errorf("Expected amount %v, got %v", req.Amount, savedTx.Amount)
	}
	if savedTx.UserID != req.UserID {
		t.Errorf("Expected userID %v, got %v", req.UserID, savedTx.UserID)
	}
	if savedTx.Type != "withdraw" {
		t.Errorf("Expected type 'withdraw', got %v", savedTx.Type)
	}
	if savedTx.Status != "pending" {
		t.Errorf("Expected status 'pending', got %v", savedTx.Status)
	}
	if savedTx.GatewayID != req.GatewayID {
		t.Errorf("Expected gatewayID %v, got %v", req.GatewayID, savedTx.GatewayID)
	}
	if savedTx.GatewayTxnId != "mock_txn_123" {
		t.Errorf("Expected gatewayTxnId 'mock_txn_123', got %v", savedTx.GatewayTxnId)
	}

}

func TestWithdraw_PaymentProcessingFailure(t *testing.T) {
	service, mockGateway, _ := setupTestService(t, true, 1000)
	mockGateway.shouldFail = true

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	result, err := service.Withdraw(req)
	if err == nil {
		t.Error("Expected error due to payment processing failure, got success")
	}
	if result != nil {
		t.Error("Expected nil result, got result")
	}
	if err != nil && err.Error() != "Payment gateway error." {
		t.Errorf("Expected payment processing error, got: %v", err)
	}

}

//----------------------------------------  Withdraw Test ----------------------------------------------------//

func TestWithdraw_InsufficientFunds(t *testing.T) {
	service, _, _ := setupTestService(t, true, 50)

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	_, err := service.Withdraw(req)
	if err == nil {
		t.Error("Expected insufficient funds error, but got success")
	}
	if err != nil && err.Error() != "Insufficient funds." {
		t.Errorf("Expected insufficient funds error, got: %v", err)
	}
}

func TestWithdraw_ComplianceFailure(t *testing.T) {
	service, _, _ := setupTestService(t, false, 1000)

	req := &models.TransactionRequest{
		Amount:    100,
		Currency:  "USD",
		GatewayID: 1,
		CountryID: 840,
		UserID:    1,
	}

	_, err := service.Withdraw(req)
	if err == nil {
		t.Error("Expected compliance check to fail, but got success")
	}
	if err != nil && err.Error() != "Compliance check failed: failed" {
		t.Errorf("Expected compliance error, got: %v", err)
	}
}

func TestHandleCallback_Success(t *testing.T) {
	service, _, mockRepo := setupTestService(t, true, 1000)

	tx := &db.Transaction{
		Amount:       100,
		Type:         "deposit",
		UserID:       1,
		GatewayID:    1,
		Status:       "pending",
		GatewayTxnId: "txn123",
	}
	mockRepo.Create(tx)

	callback := &models.PaymentCallback{
		GatewayTxnID: "txn123",
		Status:       "completed",
		GatewayID:    1,
	}

	err := service.HandleCallback(callback)
	if err != nil {
		t.Errorf("Expected successful callback handling, got error: %v", err)
	}

	updatedTx, err := mockRepo.GetTransactionByGatewayTxnId("txn123")
	if err != nil {
		t.Errorf("Failed to fetch updated transaction: %v", err)
	}
	if updatedTx == nil {
		t.Error("Transaction not found in repository")
	}
	if updatedTx.Status != db.StatusCompleted {
		t.Errorf("Expected transaction status to be 'failed', got '%s'", updatedTx.Status)
	}

}
func TestHandleCallback_Failed(t *testing.T) {
	service, _, mockRepo := setupTestService(t, true, 1000)

	tx := &db.Transaction{
		Amount:       100,
		Type:         "deposit",
		UserID:       1,
		GatewayID:    1,
		Status:       "pending",
		GatewayTxnId: "txn123",
	}
	mockRepo.Create(tx)

	callback := &models.PaymentCallback{
		GatewayTxnID: "txn123",
		Status:       "failed",
		GatewayID:    1,
	}

	err := service.HandleCallback(callback)
	if err != nil {
		t.Errorf("Expected successful callback handling, got error: %v", err)
	}

	updatedTx, err := mockRepo.GetTransactionByGatewayTxnId("txn123")
	if err != nil {
		t.Errorf("Failed to fetch updated transaction: %v", err)
	}
	if updatedTx == nil {
		t.Error("Transaction not found in repository")
	}
	if updatedTx.Status != db.StatusFailed {
		t.Errorf("Expected transaction status to be 'failed', got '%s'", updatedTx.Status)
	}
}

func TestHandleCallback_InvalidTransaction(t *testing.T) {
	service, _, _ := setupTestService(t, true, 1000)

	mockCallback := &models.PaymentCallback{
		GatewayTxnID: "invalid_txn",
		Status:       "completed",
		GatewayID:    1,
	}

	err := service.HandleCallback(mockCallback)
	if err == nil {
		t.Error("Expected error for invalid transaction, but got success")
	}
	if err != nil && err.Error() != "Transaction not found" {
		t.Errorf("Expected 'Transaction not found' error, got: %v", err)
	}
}
