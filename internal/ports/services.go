package ports

import "context"

type PaymentService interface {
	CreateIntent(ctx context.Context, orderID string, amount float64, currency string) (clientSecret string, err error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) error
	Refund(ctx context.Context, paymentIntentID string, amount float64) error
}

type NotificationService interface {
	OrderConfirmed(ctx context.Context, orderID string) error
	OrderStatusChanged(ctx context.Context, orderID string) error
	LowStockAlert(ctx context.Context, productID, branchID string) error
}

type EventBus interface {
	Publish(ctx context.Context, channel string, payload any) error
	Subscribe(ctx context.Context, channel string, fn func([]byte)) error
}

type CacheService interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttlSeconds int) error
	Delete(ctx context.Context, key string) error
	Invalidate(ctx context.Context, pattern string) error
}

type StorageService interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	UploadAuto(ctx context.Context, folder string, data []byte, name, ct string) (url string, err error)
	Delete(ctx context.Context, key string) error
	SignedURL(ctx context.Context, key string) (url string, err error)
}
