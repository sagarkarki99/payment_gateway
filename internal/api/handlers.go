package api

import (
	"net/http"
	"payment-gateway/internal/middleware"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"payment-gateway/internal/utils"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.NewPaymentService(),
	}
}

// @Summary Initiate a deposit
// @Description Process a deposit request with idempotency support
// @Tags Transactions
// @Accept json,application/xml
// @Produce json,application/xml
// @Param Idempotency-Key header string true "Unique key for request idempotency (UUID format)" example(123e4567-e89b-12d3-a456-426614174000)
// @Param request body models.TransactionRequest true "Deposit request details"
// @Success 200 {object} models.APIResponse{data=models.PaymentResult} "Deposit initiated successfully"
// @Failure 400 {object} models.APIError "Invalid request parameters or validation error"
// @Failure 422 {object} models.APIError "Payment processing failed"
// @Failure 500 {object} models.APIError "Internal server error"
// @Failure 502 {object} models.APIError "Payment gateway error"
// @Router /deposit [post]
func (ph *PaymentHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(int)

	req := models.TransactionRequest{
		UserID: userID,
	}
	if err := utils.DecodeRequest(r, &req); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, "Could not parse data"))
		return
	}

	if err := req.Validate(); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, err.Error()))
		return
	}

	ph.handleIdempotency(w, r, func() (*models.APIResponse, error) {
		result, err := ph.paymentService.Deposit(&req)
		if err != nil {
			return nil, err
		}
		return &models.APIResponse{
			StatusCode: http.StatusOK,
			Message:    "Deposit initiated",
			Data:       result,
		}, nil
	})
}

// @Summary Initiate a withdrawal
// @Description Process a withdrawal request with idempotency support
// @Tags Transactions
// @Accept json,application/xml
// @Produce json,application/xml
// @Param Idempotency-Key header string true "Unique key for request idempotency (UUID format)" example(123e4567-e89b-12d3-a456-426614174000)
// @Param request body models.TransactionRequest true "Withdrawal request details"
// @Success 200 {object} models.APIResponse{data=models.PaymentResult} "Withdrawal initiated successfully"
// @Failure 400 {object} models.APIError "Invalid request parameters or validation error"
// @Failure 422 {object} models.APIError "Insufficient funds or payment processing failed"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /withdraw [post]
func (ph *PaymentHandler) WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeUnauthorized, "User ID not found in context"))
		return
	}

	req := models.TransactionRequest{
		UserID: userID,
	}
	if err := utils.DecodeRequest(r, &req); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, "Could not parse data"))
		return
	}

	if err := req.Validate(); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, err.Error()))
		return
	}

	ph.handleIdempotency(w, r, func() (*models.APIResponse, error) {
		result, err := ph.paymentService.Withdraw(&req)
		if err != nil {
			return nil, err
		}
		return &models.APIResponse{
			StatusCode: http.StatusOK,
			Message:    "Withdrawal initiated",
			Data:       result,
		}, nil
	})
}

// @Summary Handle payment gateway callback
// @Description Process callback notifications from payment gateways
// @Tags Callbacks
// @Accept json,application/xml
// @Produce json,application/xml
// @Param request body models.PaymentCallback true "Callback notification details"
// @Success 200 {object} models.APIResponse "Callback processed successfully"
// @Failure 400 {object} models.APIError "Invalid callback data or validation error"
// @Failure 404 {object} models.APIError "Transaction not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /payment-callback [post]
func (ph *PaymentHandler) PaymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	gatewayID, _ := r.Context().Value(middleware.GatewayIDKey).(int)

	callback := models.PaymentCallback{
		GatewayID: gatewayID,
	}
	if err := utils.DecodeCallbackRequest(r, &callback); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, "Invalid data. It should include transaction ID and status"))
		return
	}

	if err := callback.Validate(); err != nil {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, err.Error()))
		return
	}

	if err := ph.paymentService.HandleCallback(&callback); err != nil {
		utils.WriteErrorResponse(w, r, err)
		return
	}

	utils.WriteResponse(w, r, http.StatusOK, models.APIResponse{
		StatusCode: http.StatusOK,
		Message:    "Callback processed successfully",
	})
}

func (ph *PaymentHandler) handleIdempotency(w http.ResponseWriter, r *http.Request, process func() (*models.APIResponse, error)) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		utils.WriteErrorResponse(w, r, models.NewServiceError(models.ErrorCodeValidation, "Idempotency-Key header is required"))
		return
	}

	response, err := process()
	if err != nil {
		utils.WriteErrorResponse(w, r, err)
		return
	}

	utils.WriteResponse(w, r, response.StatusCode, response)
}
