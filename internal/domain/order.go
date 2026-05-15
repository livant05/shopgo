package domain

import "time"

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
	StatusRefunded   OrderStatus = "refunded"
)

// validTransitions es inmutable — define el contrato del dominio.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:    {StatusConfirmed, StatusCancelled},
	StatusConfirmed:  {StatusProcessing, StatusCancelled},
	StatusProcessing: {StatusShipped},
	StatusShipped:    {StatusDelivered},
	StatusDelivered:  {StatusRefunded},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	for _, t := range validTransitions[s] {
		if t == next {
			return true
		}
	}
	return false
}

type Order struct {
	ID              string      `json:"id"`
	BranchID        string      `json:"branch_id"`
	CustomerID      string      `json:"customer_id"`
	Status          OrderStatus `json:"status"`
	Items           []OrderItem `json:"items"`
	Subtotal        float64     `json:"subtotal"`
	Tax             float64     `json:"tax"`
	Discount        float64     `json:"discount"`
	ShippingCost    float64     `json:"shipping_cost"`
	Total           float64     `json:"total"`
	Currency        string      `json:"currency"`
	ShippingAddress Address     `json:"shipping_address"`
	PaymentIntentID string      `json:"payment_intent_id,omitempty"`
	ReservationID   string      `json:"-"`
	CouponCode      string      `json:"coupon_code,omitempty"`
	Notes           string      `json:"notes,omitempty"`
	RefundStatus    string      `json:"refund_status"`
	RefundReason    string      `json:"refund_reason,omitempty"`
	RefundedAt      *time.Time  `json:"refunded_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unit_price"`
	Quantity  int     `json:"quantity"`
	LineTotal float64 `json:"line_total"`
}

type Inventory struct {
	ProductID    string    `json:"product_id"`
	BranchID     string    `json:"branch_id"`
	Quantity     int       `json:"quantity"`
	ReservedQty  int       `json:"reserved_qty"`
	ReorderPoint int       `json:"reorder_point"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (i *Inventory) Available() int { return i.Quantity - i.ReservedQty }
func (i *Inventory) IsLow() bool    { return i.Available() <= i.ReorderPoint }

type InventoryMovement struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductSKU     string    `json:"product_sku"`
	FromBranchID   string    `json:"from_branch_id,omitempty"`
	FromBranchName string    `json:"from_branch_name,omitempty"`
	ToBranchID     string    `json:"to_branch_id,omitempty"`
	ToBranchName   string    `json:"to_branch_name,omitempty"`
	Quantity       int       `json:"quantity"`
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Note           string    `json:"note,omitempty"`
	UserID         string    `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type Coupon struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	Type       string     `json:"type"` // "percent" | "fixed"
	Value      float64    `json:"value"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	MaxUses    *int       `json:"max_uses,omitempty"`
	UsesCount  int        `json:"uses_count"`
	IsActive   bool       `json:"is_active"`
}

func (c *Coupon) Apply(subtotal float64) float64 {
	if c.Type == "percent" {
		return subtotal * (c.Value / 100)
	}
	if c.Value > subtotal {
		return subtotal
	}
	return c.Value
}
