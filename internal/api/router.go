package api

import (
	"net/http"
	_ "payment-gateway/docs" // This line is important for swagger
	"payment-gateway/internal/middleware"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Initialize payment handler with unified service
	ph := NewPaymentHandler()

	// User authenticated routes (deposit/withdraw)
	userAPI := router.PathPrefix("").Subrouter()
	userAPI.Use(middleware.UserAuthMiddleware)
	userAPI.HandleFunc("/deposit", ph.Deposit).Methods(http.MethodPost)
	userAPI.HandleFunc("/withdraw", ph.WithdrawalHandler).Methods(http.MethodPost)

	// Gateway authenticated routes (payment callbacks)
	gatewayAPI := router.PathPrefix("").Subrouter()
	gatewayAPI.Use(middleware.GatewayAuthMiddleware)
	gatewayAPI.HandleFunc("/payment-callback", ph.PaymentCallbackHandler).Methods(http.MethodPost)

	return router
}
