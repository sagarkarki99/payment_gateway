package services

import (
	"payment-gateway/internal/models"
	"time"
)

type ComplianceService interface {
	CheckStatus(req *models.TransactionRequest) (string, error)
}

type MyComplianceService struct {
}

func (mcs *MyComplianceService) CheckStatus(req *models.TransactionRequest) (string, error) {
	// All the complianceService logic goes here...
	time.Sleep(2 * time.Second) // compliance check logic
	return "approved", nil
}
