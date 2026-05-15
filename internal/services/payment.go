package services

import (
	"context"
	"fmt"
	"log/slog"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	stripewebhook "github.com/stripe/stripe-go/v76/webhook"
	"github.com/yourorg/shopgo/internal/ports"
)

// PaymentService — Stripe DIRECTO.
// Sin Connect, sin ApplicationFee, sin plataforma overhead.
// El 100% del pago va a tu cuenta Stripe.
type PaymentService struct {
	orders        ports.OrderRepository
	cache         ports.CacheService
	webhookSecret string
}

func NewPaymentService(orders ports.OrderRepository, cache ports.CacheService, stripeKey, webhookSecret string) *PaymentService {
	stripe.Key = stripeKey
	return &PaymentService{orders: orders, cache: cache, webhookSecret: webhookSecret}
}

func (s *PaymentService) CreateIntent(ctx context.Context, orderID string, amount float64, currency string) (string, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{"order_id": orderID},
	}
	pi, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("crear payment intent: %w", err)
	}
	return pi.ClientSecret, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, sig string) error {
	event, err := stripewebhook.ConstructEvent(payload, sig, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("firma webhook inválida: %w", err)
	}

	// Idempotencia: 72h
	cacheKey := "stripe:" + event.ID
	var seen bool
	if s.cache.Get(ctx, cacheKey, &seen) == nil {
		slog.Info("evento Stripe ya procesado", "id", event.ID)
		return nil
	}

	var handlerErr error
	switch event.Type {
	case "payment_intent.succeeded":
		handlerErr = s.onSuccess(ctx, event)
	case "payment_intent.payment_failed":
		handlerErr = s.onFailed(ctx, event)
	default:
		slog.Debug("evento Stripe ignorado", "type", event.Type)
	}

	if handlerErr == nil {
		s.cache.Set(ctx, cacheKey, true, 72*3600)
	}
	return handlerErr
}

func (s *PaymentService) onSuccess(ctx context.Context, event stripe.Event) error {
	meta := event.Data.Object["metadata"].(map[string]any)
	orderID := meta["order_id"].(string)
	piID := event.Data.Object["id"].(string)
	if err := s.orders.ConfirmPayment(ctx, orderID, piID); err != nil {
		return fmt.Errorf("confirmar orden %s: %w", orderID, err)
	}
	slog.Info("pago confirmado", "order_id", orderID)
	return nil
}

func (s *PaymentService) onFailed(ctx context.Context, event stripe.Event) error {
	meta := event.Data.Object["metadata"].(map[string]any)
	orderID := meta["order_id"].(string)
	if err := s.orders.UpdateStatus(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("cancelar orden %s: %w", orderID, err)
	}
	slog.Warn("pago fallido, orden cancelada", "order_id", orderID)
	return nil
}
