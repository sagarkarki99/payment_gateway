package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	GatewayIDKey contextKey = "gateway_id"
)

// This middleware is used to authorize User.
func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// We should get user information from UserService or any other relevant service.
		// For now, I will just put a static userID.

		userId := 33322

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// This middleware is used to authorize payment gateway.
func GatewayAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// apiKey := r.Header.Get("X-Gateway-API-Key")
		// if apiKey == "" {
		// 	utils.WriteResponse(w, r, http.StatusUnauthorized, models.APIResponse{
		// 		StatusCode: http.StatusUnauthorized,
		// 		Message:    "Gateway API key is required",
		// 	})
		// 	return
		// }

		// We should get gateway ID from GatewayService or any other relevant source.
		// For now, I will just put a static gatewayID.
		gatewayId := 433434

		// Add gateway ID to request context
		ctx := context.WithValue(r.Context(), GatewayIDKey, gatewayId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
