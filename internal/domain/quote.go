package domain

import "time"

type QuoteItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type Quote struct {
	ID            string      `json:"id"`
	QuoteNumber   int         `json:"quote_number"`
	Items         []QuoteItem `json:"items"`
	Subtotal      float64     `json:"subtotal"`
	TaxRate       float64     `json:"tax_rate"`
	TaxAmount     float64     `json:"tax_amount"`
	Total         float64     `json:"total"`
	Currency      string      `json:"currency"`
	StoreName     string      `json:"store_name"`
	ContactEmail  string      `json:"contact_email"`
	SupportPhone  string      `json:"support_phone"`
	CustomerName  string      `json:"customer_name"`
	CustomerEmail string      `json:"customer_email"`
	CustomerPhone string      `json:"customer_phone"`
	Note          string      `json:"note"`
	CreatedAt     time.Time   `json:"created_at"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
}
