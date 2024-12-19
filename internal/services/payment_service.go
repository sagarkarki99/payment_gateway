package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"payment-gateway/db"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/security"
	"payment-gateway/internal/utils"
)

type PaymentService interface {
	// This will deposit the amount to user account.
	Deposit(req *models.TransactionRequest) (*models.PaymentResult, error)

	// This will will withdraw from user amount.
	Withdraw(req *models.TransactionRequest) (*models.PaymentResult, error)

	// This function is for external payment gateway to confirm any transaction.
	HandleCallback(callbackData *models.PaymentCallback) error
}

type paymentService struct {
	cs   ComplianceService
	as   AccountService
	repo db.TransactionRepository
}

func NewPaymentService() PaymentService {
	return &paymentService{
		cs:   &MyComplianceService{},
		as:   NewAccountService(),
		repo: db.NewTransactionRepository(db.Db),
	}
}

func (p *paymentService) Deposit(req *models.TransactionRequest) (*models.PaymentResult, error) {
	if _, err := p.cs.CheckStatus(req); err != nil {
		return nil, models.NewServiceError(models.ErrorCodeValidation, err.Error())
	}

	trx := &db.Transaction{
		Amount:    req.Amount,
		Type:      db.TypeDeposit,
		UserID:    req.UserID,
		GatewayID: req.GatewayID,
		CreatedAt: time.Now(),
		CountryID: req.CountryID,
	}

	err := p.processTransaction(trx)
	if err != nil {
		return nil, err
	}

	return &models.PaymentResult{
		TransactionId: trx.ID,
	}, nil
}

func (p *paymentService) Withdraw(req *models.TransactionRequest) (*models.PaymentResult, error) {
	// Get balance from account service
	balance, err := p.as.GetBalance(req.UserID)
	if err != nil {
		return nil, models.NewServiceError(models.ErrorCodeUnknown, "Failed to get account balance: "+err.Error())
	}

	if err = p.validateBalance(balance, req); err != nil {
		return nil, err
	}

	// Compliance check after balance validation
	if sts, err := p.cs.CheckStatus(req); err != nil {
		return nil, models.NewServiceError(models.ErrorCodeValidation, "Compliance check failed: "+sts)
	}

	trx := &db.Transaction{
		Amount:    req.Amount,
		Type:      db.TypeWithdraw,
		UserID:    req.UserID,
		GatewayID: req.GatewayID,
		CreatedAt: time.Now(),
		CountryID: req.CountryID,
	}

	err = p.processTransaction(trx)
	if err != nil {
		return nil, err
	}

	return &models.PaymentResult{
		TransactionId: trx.ID,
	}, nil
}

func (p *paymentService) HandleCallback(callbackData *models.PaymentCallback) error {
	// Fetch the original transaction
	trx, err := p.repo.GetTransactionByGatewayTxnId(callbackData.GatewayTxnID)
	if err != nil {
		return models.NewServiceError(models.ErrorCodeUnknown, "Failed to fetch transaction: "+err.Error())
	}
	if trx == nil {
		return models.NewServiceError(models.ErrorCodeNotFound, "Transaction not found")
	}

	if transactionAlreadyProcessed(trx, callbackData) {
		// Ignore the status update because we have already processed this transaction.
		return nil
	}

	// Update transaction status based on what we received in the callback.
	trx.Status = callbackData.Status
	if err := p.repo.Update(*trx); err != nil {
		return models.NewServiceError(models.ErrorCodeUnknown, "Failed to update transaction.")
	}

	// Publish status update to Kafka
	go SendToKafka(trx)

	return nil
}

func SendToKafka(trx *db.Transaction) {
	jsonMsg, _ := json.Marshal(map[string]interface{}{
		"status": trx.Status,
		"userId": security.MaskData([]byte(fmt.Sprint(trx.UserID))),
		"amount": security.MaskData([]byte(fmt.Sprintf("%.2f", trx.Amount))),
		"type":   trx.Type,
	})

	err := utils.PublishWithCircuitBreaker(func() error {
		return kafka.PublishTransaction(context.Background(), fmt.Sprint(trx.ID), jsonMsg, "application/json")
	})
	if err != nil {
		// Handling this error is critical because the transaction is already recorded but
		// failed to publish to the topic for other services. In this case, we can apply couple of approaches.
		// 1. Send it to different topic such as Dead Letter Queue (DLQ) for manual intervention.
		// 2. Or, we save the failed event in database and try to send it again after some time
		// using cronjob and clear it.
		log.Printf("failed to publish callback to kafka: %v", err)
	}
}

func transactionAlreadyProcessed(trx *db.Transaction, callbackData *models.PaymentCallback) bool {
	return trx.Status == callbackData.Status
}

func (p *paymentService) processTransaction(trx *db.Transaction) error {
	gt := GetPaymentGateway(trx.CountryID, trx.GatewayID)

	err := utils.RetryOperation(func() error {
		// Create a new background context for the critical section.
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		result, err := gt.ProcessPayment(ctx, trx)
		if err != nil {
			return err
		}
		trx.Status = db.StatusPending
		trx.GatewayTxnId = result.GatewayTxnId
		return nil
	}, 3)
	if err != nil {
		return models.NewServiceError(models.ErrorCodeGatewayError, "Payment gateway error.")
	}

	savedTrx, err := p.repo.Create(trx)
	if err != nil {
		return models.NewServiceError(models.ErrorCodeUnknown, "Failed to save transaction.")
	}

	trx.ID = savedTrx.ID

	go SendToKafka(trx)

	return nil
}

func (p *paymentService) validateBalance(balance float64, req *models.TransactionRequest) error {
	if balance < req.Amount {
		return models.NewServiceError(
			models.ErrorCodeInsufficientFunds,
			"Insufficient funds.",
		)
	}
	// It will consist of bunch of other validations like daily limit, weekly limit,
	// minimum amount for withdrawal and so on.
	return nil
}
