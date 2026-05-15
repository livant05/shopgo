package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
	"github.com/yourorg/shopgo/internal/services"
)

// ── Mocks ─────────────────────────────────────────────────

type mockOrders struct {
	createFn func(context.Context, *domain.Order) (*domain.Order, error)
	updateFn func(context.Context, string, domain.OrderStatus) error
	getFn    func(context.Context, string) (*domain.Order, error)
}

func (m *mockOrders) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &domain.Order{Status: domain.StatusPending}, nil
}
func (m *mockOrders) List(_ context.Context, _ ports.OrderFilter) (*ports.Page[domain.Order], error) {
	return &ports.Page[domain.Order]{}, nil
}
func (m *mockOrders) Create(ctx context.Context, o *domain.Order) (*domain.Order, error) {
	return m.createFn(ctx, o)
}
func (m *mockOrders) UpdateStatus(ctx context.Context, id string, s domain.OrderStatus) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, s)
	}
	return nil
}
func (m *mockOrders) ConfirmPayment(_ context.Context, _, _ string) error { return nil }

type mockInventory struct {
	reserveOK bool
	released  bool
}

func (m *mockInventory) Get(_ context.Context, _, _ string) (*domain.Inventory, error) {
	return &domain.Inventory{}, nil
}
func (m *mockInventory) List(_ context.Context, _ string) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInventory) GetLowStock(_ context.Context, _ string) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInventory) Adjust(_ context.Context, _, _ string, _ int, _, _, _ string) (*domain.Inventory, error) {
	return &domain.Inventory{}, nil
}
func (m *mockInventory) Reserve(_ context.Context, _ string, _ []ports.ReserveItem) (string, error) {
	if !m.reserveOK {
		return "", errors.New("sin stock")
	}
	return "res-1", nil
}
func (m *mockInventory) Commit(_ context.Context, _ string) error              { return nil }
func (m *mockInventory) Release(_ context.Context, _ string) error             { m.released = true; return nil }
func (m *mockInventory) Transfer(_ context.Context, _ ports.TransferCmd) error { return nil }

type mockProducts struct{}

func (m *mockProducts) GetByID(_ context.Context, _ string) (*domain.Product, error) {
	return &domain.Product{BasePrice: 100}, nil
}
func (m *mockProducts) GetWithPrice(_ context.Context, _, _ string) (*domain.Product, error) {
	p := 100.0
	return &domain.Product{SKU: "T01", Name: "Producto Test", BasePrice: 100, BranchPrice: &p}, nil
}
func (m *mockProducts) List(_ context.Context, _ domain.ProductFilter) (*ports.Page[domain.Product], error) {
	return nil, nil
}
func (m *mockProducts) Create(_ context.Context, p *domain.Product) (*domain.Product, error) {
	return p, nil
}
func (m *mockProducts) Update(_ context.Context, p *domain.Product) (*domain.Product, error) {
	return p, nil
}
func (m *mockProducts) SetActive(_ context.Context, _ string, _ bool) error            { return nil }
func (m *mockProducts) SetBranchPrice(_ context.Context, _ domain.OverridePrice) error { return nil }
func (m *mockProducts) BulkUpsert(_ context.Context, _ []*domain.Product) (int, error) { return 0, nil }
func (m *mockProducts) ListCategories(_ context.Context) ([]domain.Category, error)    { return nil, nil }

type mockBranches struct{ active bool }

func (m *mockBranches) GetByID(_ context.Context, _ string) (*domain.Branch, error) {
	return &domain.Branch{IsActive: m.active, Settings: domain.BranchSettings{TaxRate: 0.16, Currency: "MXN"}}, nil
}
func (m *mockBranches) List(_ context.Context) ([]*domain.Branch, error) { return nil, nil }
func (m *mockBranches) Create(_ context.Context, b *domain.Branch) (*domain.Branch, error) {
	return b, nil
}
func (m *mockBranches) Update(_ context.Context, b *domain.Branch) (*domain.Branch, error) {
	return b, nil
}
func (m *mockBranches) SetActive(_ context.Context, _ string, _ bool) error { return nil }

type mockBus struct{}

func (m *mockBus) Publish(_ context.Context, _ string, _ any) error            { return nil }
func (m *mockBus) Subscribe(_ context.Context, _ string, _ func([]byte)) error { return nil }

// ── Tests ─────────────────────────────────────────────────

func newSvc(orders *mockOrders, inv *mockInventory, branchActive bool) *services.OrderService {
	return services.NewOrderService(orders, inv, &mockProducts{}, &mockBranches{active: branchActive}, &mockBus{})
}

func TestCreate_Success(t *testing.T) {
	inv := &mockInventory{reserveOK: true}
	orders := &mockOrders{
		createFn: func(_ context.Context, o *domain.Order) (*domain.Order, error) {
			o.ID = "order-001"
			return o, nil
		},
	}
	svc := newSvc(orders, inv, true)

	order, err := svc.Create(context.Background(), services.CreateOrderInput{
		BranchID:   "branch-1",
		CustomerID: "cust-1",
		Items:      []services.OrderItemInput{{ProductID: "prod-1", Quantity: 2}},
	})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if order.ID == "" {
		t.Error("ID vacío")
	}
	if order.Total <= 0 {
		t.Error("total debe ser > 0")
	}
	if order.Tax <= 0 {
		t.Error("IVA debe ser > 0")
	}
}

func TestCreate_BranchInactive(t *testing.T) {
	orders := &mockOrders{createFn: func(_ context.Context, o *domain.Order) (*domain.Order, error) { return o, nil }}
	svc := newSvc(orders, &mockInventory{reserveOK: true}, false)

	_, err := svc.Create(context.Background(), services.CreateOrderInput{
		BranchID: "b", Items: []services.OrderItemInput{{ProductID: "p", Quantity: 1}},
	})

	if !errors.Is(err, services.ErrBranchUnavailable) {
		t.Errorf("esperaba ErrBranchUnavailable, obtuvo: %v", err)
	}
}

func TestCreate_InsufficientStock(t *testing.T) {
	orders := &mockOrders{createFn: func(_ context.Context, o *domain.Order) (*domain.Order, error) { return o, nil }}
	svc := newSvc(orders, &mockInventory{reserveOK: false}, true)

	_, err := svc.Create(context.Background(), services.CreateOrderInput{
		BranchID: "b", Items: []services.OrderItemInput{{ProductID: "p", Quantity: 99}},
	})

	if !errors.Is(err, ports.ErrInsufficientStock) {
		t.Errorf("esperaba ErrInsufficientStock, obtuvo: %v", err)
	}
}

func TestCreate_DBError_ReleasesStock(t *testing.T) {
	inv := &mockInventory{reserveOK: true}
	orders := &mockOrders{
		createFn: func(_ context.Context, _ *domain.Order) (*domain.Order, error) {
			return nil, errors.New("connection refused")
		},
	}
	svc := newSvc(orders, inv, true)

	_, err := svc.Create(context.Background(), services.CreateOrderInput{
		BranchID: "b", Items: []services.OrderItemInput{{ProductID: "p", Quantity: 1}},
	})

	if err == nil {
		t.Fatal("debía fallar")
	}
	if !inv.released {
		t.Error("Release() debe llamarse en rollback")
	}
}

func TestCreate_PriceSnapshot(t *testing.T) {
	inv := &mockInventory{reserveOK: true}
	var captured *domain.Order
	orders := &mockOrders{
		createFn: func(_ context.Context, o *domain.Order) (*domain.Order, error) {
			captured = o
			o.ID = "x"
			return o, nil
		},
	}
	svc := newSvc(orders, inv, true)
	svc.Create(context.Background(), services.CreateOrderInput{
		BranchID: "b", Items: []services.OrderItemInput{{ProductID: "p", Quantity: 3}},
	})

	if captured == nil {
		t.Fatal("orden no capturada")
	}
	if captured.Items[0].UnitPrice != 100 {
		t.Errorf("precio esperado 100, obtuvo %.2f", captured.Items[0].UnitPrice)
	}
	if captured.Items[0].LineTotal != 300 {
		t.Errorf("line_total esperado 300, obtuvo %.2f", captured.Items[0].LineTotal)
	}
}

func TestUpdateStatus_ValidTransition(t *testing.T) {
	svc := newSvc(&mockOrders{}, &mockInventory{}, true)
	if err := svc.UpdateStatus(context.Background(), "o1", "confirmed"); err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

func TestUpdateStatus_InvalidTransition(t *testing.T) {
	svc := newSvc(&mockOrders{}, &mockInventory{}, true)
	err := svc.UpdateStatus(context.Background(), "o1", "shipped") // pending → shipped es inválido
	if !errors.Is(err, ports.ErrInvalidTransition) {
		t.Errorf("esperaba ErrInvalidTransition, obtuvo: %v", err)
	}
}

func TestOrderStatus_Transitions(t *testing.T) {
	cases := []struct {
		from   domain.OrderStatus
		to     domain.OrderStatus
		expect bool
	}{
		{domain.StatusPending, domain.StatusConfirmed, true},
		{domain.StatusPending, domain.StatusCancelled, true},
		{domain.StatusPending, domain.StatusShipped, false},
		{domain.StatusConfirmed, domain.StatusProcessing, true},
		{domain.StatusProcessing, domain.StatusShipped, true},
		{domain.StatusShipped, domain.StatusDelivered, true},
		{domain.StatusDelivered, domain.StatusShipped, false},
		{domain.StatusCancelled, domain.StatusPending, false},
	}

	for _, tc := range cases {
		got := tc.from.CanTransitionTo(tc.to)
		if got != tc.expect {
			t.Errorf("%s → %s: esperaba %v, obtuvo %v", tc.from, tc.to, tc.expect, got)
		}
	}
}
