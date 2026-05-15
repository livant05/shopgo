package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	stripego "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
)

type orderConfirmer interface {
	ConfirmPayment(ctx context.Context, orderID, paymentIntentID string) error
}

type PaymentService struct {
	webhookSecret string
	orders        orderConfirmer
}

func NewPaymentService(secretKey, webhookSecret string, orders orderConfirmer) *PaymentService {
	stripego.Key = secretKey
	return &PaymentService{webhookSecret: webhookSecret, orders: orders}
}

func (s *PaymentService) CreateIntent(_ context.Context, orderID string, amount float64, currency string) (string, error) {
	params := &stripego.PaymentIntentParams{
		Amount:   stripego.Int64(int64(amount * 100)),
		Currency: stripego.String(currency),
		Metadata: map[string]string{"order_id": orderID},
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe create intent: %w", err)
	}
	return pi.ClientSecret, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("stripe webhook: %w", err)
	}
	if event.Type == "payment_intent.succeeded" {
		var pi stripego.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return err
		}
		orderID := pi.Metadata["order_id"]
		if orderID == "" {
			return nil
		}
		return s.orders.ConfirmPayment(ctx, orderID, pi.ID)
	}
	return nil
}

func (s *PaymentService) Refund(_ context.Context, paymentIntentID string, amount float64) error {
	_, err := refund.New(&stripego.RefundParams{
		PaymentIntent: stripego.String(paymentIntentID),
		Amount:        stripego.Int64(int64(amount * 100)),
	})
	return err
}
