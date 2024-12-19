package models

import "fmt"

// @title Payment Gateway API
// @version 1.0
// @description A payment gateway service that handles deposits and withdrawals with idempotency support
// @host localhost:8000
// @BasePath /
// @schemes http https
// @accept json xml
// @produce json xml
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// TransactionRequest represents the request payload for transactions
// @Description Transaction request model
type TransactionRequest struct {
	// Amount to process
	// required: true
	Amount float64 `json:"amount" xml:"amount" example:"99.99"`
	// Currency code in ISO 4217 format
	// required: true
	Currency string `json:"currency" xml:"currency" example:"USD"`
	// Payment gateway identifier
	// required: true
	GatewayID int `json:"gateway_id" xml:"gateway_id" example:"112"`
	// Country identifier (ISO 3166-1 numeric)
	// required: true
	CountryID int `json:"country_id" xml:"country_id" example:"840"`

	// Internal field, not exposed in swagger
	UserID int `json:"user_id" xml:"user_id" swaggerignore:"true"`
}

func (t *TransactionRequest) Validate() error {
	if t.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	} else if t.Currency == "" || len(t.Currency) > 3 {
		return fmt.Errorf("invalid currency code")
	} else if t.GatewayID <= 0 {
		return fmt.Errorf("invalid gateway id")
	} else if t.CountryID <= 0 {
		return fmt.Errorf("invalid country id")
	}
	return nil
}

// APIResponse represents the successful API response structure
// @Description Successful API response model
type APIResponse struct {
	// HTTP status code
	// required: true
	StatusCode int `json:"status_code" xml:"status_code" example:"200"`
	// Response message
	// required: true
	Message string `json:"message" xml:"message" example:"Operation successful"`
	// Response data (only present in successful responses)
	// required: false
	Data interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// APIError represents the error response structure
// @Description Error response model
type APIError struct {
	// HTTP status code
	// required: true
	StatusCode int `json:"status_code" xml:"status_code" example:"400"`
	// Error message
	// required: true
	Error string `json:"error" xml:"error" example:"Invalid request parameters"`
}

// PaymentCallback represents the callback request from payment gateways
// @Description Payment gateway callback model
type PaymentCallback struct {
	// Transaction identifier
	// required: true
	GatewayTxnID string `json:"gateway_txn_id" xml:"gateway_txn_id" example:"123456"`
	// Transaction status
	// required: true
	Status string `json:"status" xml:"status" example:"completed"`
	// Optional error message
	// required: false
	ErrorMessage string `json:"error_message,omitempty" xml:"error_message,omitempty" example:"Transaction proceeded successfully."`
	// Internal field, not exposed in swagger
	GatewayID int `json:"gateway_id" xml:"gateway_id" swaggerignore:"true"`
}

func (pc *PaymentCallback) Validate() error {
	if pc.GatewayTxnID == "" {
		return fmt.Errorf("invalid transaction id")
	}
	if pc.Status == "" {
		return fmt.Errorf("status is required")
	}
	if pc.GatewayID <= 0 {
		return fmt.Errorf("invalid gateway id")
	}
	return nil
}

// PaymentResult represents the result of a payment transaction
// @Description Payment transaction result model
type PaymentResult struct {
	// Transaction identifier
	// required: true
	TransactionId int `json:"transaction_id" xml:"transaction_id" example:"123456"`
}
