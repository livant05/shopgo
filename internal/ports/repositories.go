package ports

import (
	"context"
	"errors"

	"github.com/yourorg/shopgo/internal/domain"
)

var (
	ErrNotFound          = errors.New("registro no encontrado")
	ErrConflict          = errors.New("ya existe un registro con ese identificador")
	ErrInsufficientStock = errors.New("stock insuficiente")
	ErrInvalidTransition = errors.New("transición de estado inválida")
)

type Page[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type ReserveItem struct {
	ProductID string
	Quantity  int
}

type TransferCmd struct {
	ProductID    string
	FromBranchID string
	ToBranchID   string
	Quantity     int
	Note         string
	UserID       string
}

// — Repositorios ——————————————————————————————————————

type BranchRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Branch, error)
	List(ctx context.Context) ([]*domain.Branch, error)
	Create(ctx context.Context, b *domain.Branch) (*domain.Branch, error)
	Update(ctx context.Context, b *domain.Branch) (*domain.Branch, error)
	SetActive(ctx context.Context, id string, active bool) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) (*domain.User, error)
	UpdateProfile(ctx context.Context, id, firstName, lastName, phone string, addr domain.Address) (*domain.User, error)
	ChangePassword(ctx context.Context, id, newHash string) error
	SetActive(ctx context.Context, id string, active bool) error
	SetMFASecret(ctx context.Context, id, secret string) error
	List(ctx context.Context, page, pageSize int) (*Page[domain.User], error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetWithPrice(ctx context.Context, id, branchID string) (*domain.Product, error)
	List(ctx context.Context, f domain.ProductFilter) (*Page[domain.Product], error)
	ListTags(ctx context.Context) ([]string, error)
	Create(ctx context.Context, p *domain.Product) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) (*domain.Product, error)
	SetActive(ctx context.Context, id string, active bool) error
	SetBranchPrice(ctx context.Context, op domain.OverridePrice) error
	BulkUpsert(ctx context.Context, products []*domain.Product) (int, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id string) (*domain.Category, error)
	CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error)
	UpdateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error)
	SetCategoryActive(ctx context.Context, id string, active bool) error
}

type InventoryRepository interface {
	Get(ctx context.Context, productID, branchID string) (*domain.Inventory, error)
	List(ctx context.Context, branchID string) ([]*domain.Inventory, error)
	GetLowStock(ctx context.Context, branchID string) ([]*domain.Inventory, error)
	Adjust(ctx context.Context, productID, branchID string, delta int, reason, note, userID string) (*domain.Inventory, error)
	Restore(ctx context.Context, productID, branchID string, qty int, reason, userID string) error
	Reserve(ctx context.Context, branchID string, items []ReserveItem) (reservationID string, err error)
	Commit(ctx context.Context, reservationID string) error
	Release(ctx context.Context, reservationID string) error
	Transfer(ctx context.Context, cmd TransferCmd) error
	History(ctx context.Context, branchID, movType, from, to string, page, pageSize int) ([]*domain.InventoryMovement, int, error)
}

type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	List(ctx context.Context, f OrderFilter) (*Page[domain.Order], error)
	Create(ctx context.Context, o *domain.Order) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	ConfirmPayment(ctx context.Context, orderID, paymentIntentID string) error
	RequestRefund(ctx context.Context, orderID, reason string) error
	ApproveRefund(ctx context.Context, orderID string) error
	RejectRefund(ctx context.Context, orderID string) error
}

type OrderFilter struct {
	BranchID     string
	CustomerID   string
	Status       string
	RefundStatus string
	From         string
	To           string
	Page         int
	PageSize     int
}

type QuoteFilter struct {
	Search   string
	Status   string
	From     string
	To       string
	Page     int
	PageSize int
}

type QuoteStats struct {
	Pending        int64   `json:"pending"`
	Accepted       int64   `json:"accepted"`
	Rejected       int64   `json:"rejected"`
	Total          int64   `json:"total"`
	AcceptedValue  float64 `json:"accepted_value"`
	PipelineValue  float64 `json:"pipeline_value"`
	ConversionRate float64 `json:"conversion_rate"`
}

type QuoteRepository interface {
	Create(ctx context.Context, q *domain.Quote) (*domain.Quote, error)
	GetByID(ctx context.Context, id string) (*domain.Quote, error)
	UpdateStatus(ctx context.Context, id, status, note string) (*domain.Quote, error)
	UpdateItems(ctx context.Context, id string, items []domain.QuoteItem, subtotal, taxAmount, total float64) (*domain.Quote, error)
	ExpireOverdue(ctx context.Context) (int, error)
	Stats(ctx context.Context) (*QuoteStats, error)
	List(ctx context.Context, f QuoteFilter) (*Page[domain.Quote], error)
}

type ReportRepository interface {
	Revenue(ctx context.Context, from, to string) (*RevenueMetrics, error)
	SalesByBranch(ctx context.Context, from, to string) ([]*BranchSales, error)
	TopProducts(ctx context.Context, branchID, from, to string, n int) ([]*TopProduct, error)
	TopCustomers(ctx context.Context, from, to string, n int) ([]*TopCustomer, error)
	HourlySeries(ctx context.Context, branchID, from, to string) ([]*HourlyStat, error)
	DailySeries(ctx context.Context, branchID, from, to string) ([]*DailyStat, error)
	InventoryReport(ctx context.Context, branchID string) ([]*InvRow, error)
}

type RevenueMetrics struct {
	GMV       float64 `json:"gmv"`
	Orders    int64   `json:"orders"`
	AOV       float64 `json:"aov"`
	Customers int64   `json:"customers"`
}

type BranchSales struct {
	BranchID   string  `json:"branch_id"`
	BranchName string  `json:"branch_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
	Customers  int64   `json:"customers"`
}

type TopProduct struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	UnitsSold int64   `json:"units_sold"`
	Revenue   float64 `json:"revenue"`
}

type TopCustomer struct {
	CustomerID string  `json:"customer_id"`
	Email      string  `json:"email"`
	FullName   string  `json:"full_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
}

type HourlyStat struct {
	Hour    int     `json:"hour"`
	Orders  int64   `json:"orders"`
	Revenue float64 `json:"revenue"`
}

type DailyStat struct {
	Day     string  `json:"day"`
	Orders  int64   `json:"orders"`
	Revenue float64 `json:"revenue"`
}

type InvRow struct {
	SKU          string  `json:"sku"`
	ProductName  string  `json:"product_name"`
	BranchName   string  `json:"branch_name"`
	Quantity     int     `json:"quantity"`
	ReservedQty  int     `json:"reserved_qty"`
	Available    int     `json:"available"`
	ReorderPoint int     `json:"reorder_point"`
	IsLow        bool    `json:"is_low"`
	Price        float64 `json:"price"`
}
