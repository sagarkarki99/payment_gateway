package services

import (
	"context"
	"fmt"
	"payment-gateway/db"
	"time"
)

type GatewayResult struct {
	// This is a unique id that is returned from the gateway.
	GatewayTxnId string
	//... other necessary response parameters
}

type PaymentGateway interface {
	// This request transaction can be converted based on what Payment processor takes in.
	ProcessPayment(ctx context.Context, req *db.Transaction) (*GatewayResult, error)
}

// Make GetPaymentGateway a variable so it can be mocked in tests
var GetPaymentGateway = func(countryId int, gatewayId int) PaymentGateway {
	repo := db.NewGatewayRepository(db.Db)

	// Get all available gateways for the country
	gateways, err := repo.GetAvailableGateways(countryId)
	if err != nil {
		return &StripeGateway{}
	}

	for _, gateway := range gateways {
		if gateway.ID == gatewayId {
			return getGatewayImplementation(gateway.Name)
		}
	}

	if len(gateways) > 0 {
		// Getting the default gateway.
		return getGatewayImplementation(gateways[0].Name)
	}

	// If no gateways available, fallback to Stripe.
	// If we do not want to do this we can simply return error from here.
	return &StripeGateway{}
}

func getGatewayImplementation(gatewayName string) PaymentGateway {
	switch gatewayName {
	case "stripe":
		return &StripeGateway{}
	case "paypal":
		// return &PayPalGateway{}
		return &PaypalGateway{} // Fallback to Stripe for now
	default:
		return &StripeGateway{} // Default to Stripe
	}
}

type StripeGateway struct{}

func (stripe *StripeGateway) ProcessPayment(ctx context.Context, req *db.Transaction) (*GatewayResult, error) {
	// process payment logic for stripe. This could be an api call with stripe
	// related config.

	time.Sleep(1 * time.Second) // simulating payment logic
	randomId := "stripe_txn_" + time.Now().Format("20060102150405")
	fmt.Println("Stripe payment processed with txn id: ", randomId)
	return &GatewayResult{
		GatewayTxnId: randomId,
	}, nil
}

type PaypalGateway struct{}

func (stripe *PaypalGateway) ProcessPayment(ctx context.Context, req *db.Transaction) (*GatewayResult, error) {

	// process payment logic for paypal. This could be an api call with paypal related config.
	time.Sleep(1 * time.Second) // simulating payment logic
	return &GatewayResult{
		GatewayTxnId: "paypal_txn_id_322323",
	}, nil
}

// Can have more implementation of Gateway interface like Revolut etc.
