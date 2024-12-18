package api

import (
	"net/http"
)

// DepositHandler handles deposit requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	    "currency": "EUR"
//	}
func DepositHandler(w http.ResponseWriter, r *http.Request) {
	// deposit request logic
}

// WithdrawalHandler handles withdrawal requests (feel free to update how user is passed to the request)
// Sample Request (POST /deposit):
//
//	{
//	    "amount": 100.00,
//	    "user_id": 1,
//	    "currency": "EUR"
//	}
func WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	// withdrawal request logic
}
