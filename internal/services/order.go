package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

var ErrBranchUnavailable = errors.New("sucursal no disponible")

type OrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderInput struct {
	BranchID        string
	CustomerID      string
	Items           []OrderItemInput
	CouponCode      string
	ShippingAddress domain.Address
	Currency        string
	Notes           string
}

type OrderService struct {
	orders    ports.OrderRepository
	inventory ports.InventoryRepository
	products  ports.ProductRepository
	branches  ports.BranchRepository
	events    ports.EventBus
}

func NewOrderService(o ports.OrderRepository, i ports.InventoryRepository, p ports.ProductRepository, b ports.BranchRepository, ev ports.EventBus) *OrderService {
	return &OrderService{orders: o, inventory: i, products: p, branches: b, events: ev}
}

// Create implementa la Saga de creación: 6 pasos atómicos con rollback.
func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	// PASO 1 — Verificar sucursal activa
	branch, err := s.branches.GetByID(ctx, in.BranchID)
	if err != nil || !branch.IsActive {
		return nil, ErrBranchUnavailable
	}

	// PASO 2 — Snapshot de precios (precio congelado al momento de compra)
	items := make([]domain.OrderItem, 0, len(in.Items))
	reserveItems := make([]ports.ReserveItem, 0, len(in.Items))
	var subtotal float64

	for _, ci := range in.Items {
		p, err := s.products.GetWithPrice(ctx, ci.ProductID, in.BranchID)
		if err != nil {
			return nil, fmt.Errorf("producto %s: %w", ci.ProductID, err)
		}
		price := p.EffectivePrice()
		items = append(items, domain.OrderItem{
			ProductID: ci.ProductID,
			SKU:       p.SKU,
			Name:      p.Name,
			UnitPrice: price,
			Quantity:  ci.Quantity,
			LineTotal: price * float64(ci.Quantity),
		})
		subtotal += price * float64(ci.Quantity)
		reserveItems = append(reserveItems, ports.ReserveItem{ProductID: ci.ProductID, Quantity: ci.Quantity})
	}

	// PASO 3 — Calcular IVA (usando la tasa de la sucursal)
	taxRate := branch.Settings.TaxRate
	if taxRate == 0 {
		taxRate = 0.16
	}
	tax := subtotal * taxRate
	currency := in.Currency
	if currency == "" {
		currency = branch.Settings.Currency
	}
	if currency == "" {
		currency = "MXN"
	}

	// PASO 4 — Reservar stock (UPDATE atómico sin deadlocks)
	reservationID, err := s.inventory.Reserve(ctx, in.BranchID, reserveItems)
	if err != nil {
		return nil, ports.ErrInsufficientStock
	}

	// PASO 5 — Persistir la orden en BD
	order := &domain.Order{
		BranchID:        in.BranchID,
		CustomerID:      in.CustomerID,
		Status:          domain.StatusPending,
		Items:           items,
		Subtotal:        subtotal,
		Tax:             tax,
		Total:           subtotal + tax,
		Currency:        currency,
		ShippingAddress: in.ShippingAddress,
		ReservationID:   reservationID,
		Notes:           in.Notes,
	}

	created, err := s.orders.Create(ctx, order)
	if err != nil {
		// Rollback: liberar stock reservado
		if rbErr := s.inventory.Release(ctx, reservationID); rbErr != nil {
			slog.Error("fallo liberando reserva en rollback", "reservation_id", reservationID, "err", rbErr)
		}
		return nil, fmt.Errorf("crear orden: %w", err)
	}

	// PASO 6 — Publicar evento async (no bloquea la respuesta HTTP)
	go func() {
		payload := map[string]any{
			"event": "order.new", "order_id": created.ID, "total": created.Total,
			"branch_id": in.BranchID,
		}
		ctx := context.Background()
		if err := s.events.Publish(ctx, "orders:"+in.BranchID, payload); err != nil {
			slog.Error("publicar evento orden", "err", err)
		}
		// Fan-out al canal global que alimenta el hub SSE del admin
		if err := s.events.Publish(ctx, "notifications", payload); err != nil {
			slog.Error("publicar notificación admin", "err", err)
		}
	}()

	return created, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID, newStatus string) error {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	next := domain.OrderStatus(newStatus)
	if !order.Status.CanTransitionTo(next) {
		return ports.ErrInvalidTransition
	}
	return s.orders.UpdateStatus(ctx, orderID, next)
}

func (s *OrderService) RequestRefund(ctx context.Context, orderID, reason, customerID string) error {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.CustomerID != customerID {
		return ports.ErrNotFound
	}
	if order.Status != domain.StatusDelivered {
		return ports.ErrInvalidTransition
	}
	if order.RefundStatus != "none" {
		return ports.ErrConflict
	}
	return s.orders.RequestRefund(ctx, orderID, reason)
}

func (s *OrderService) ApproveRefund(ctx context.Context, orderID, adminID string) error {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := s.orders.ApproveRefund(ctx, orderID); err != nil {
		return err
	}
	reason := "Devolución aprobada #" + orderID[:8]
	for _, item := range order.Items {
		if restoreErr := s.inventory.Restore(ctx, item.ProductID, order.BranchID, item.Quantity, reason, adminID); restoreErr != nil {
			slog.Error("error restaurando inventario en devolución", "product_id", item.ProductID, "err", restoreErr)
		}
	}
	return nil
}

func (s *OrderService) RejectRefund(ctx context.Context, orderID string) error {
	return s.orders.RejectRefund(ctx, orderID)
}
