package postgres

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type OrderRepo struct{ db *pgxpool.Pool }

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo { return &OrderRepo{db} }

const orderSelect = `
	SELECT id, branch_id, customer_id, status, items, subtotal, tax, discount,
	       shipping_cost, total, currency, shipping_address,
	       COALESCE(payment_intent_id,''), COALESCE(coupon_code,''),
	       COALESCE(reservation_id,''), COALESCE(notes,''),
	       COALESCE(refund_status,'none'), COALESCE(refund_reason,''), refunded_at,
	       created_at, updated_at
	FROM orders`

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return r.scanOrder(r.db.QueryRow(ctx, orderSelect+` WHERE id=$1`, id))
}

func (r *OrderRepo) List(ctx context.Context, f ports.OrderFilter) (*ports.Page[domain.Order], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}

	args := []any{}
	where := []string{}
	i := 1

	if f.BranchID != "" {
		where = append(where, "branch_id=$"+argN(i))
		args = append(args, f.BranchID)
		i++
	}
	if f.CustomerID != "" {
		where = append(where, "customer_id=$"+argN(i))
		args = append(args, f.CustomerID)
		i++
	}
	if f.Status != "" {
		where = append(where, "status=$"+argN(i))
		args = append(args, f.Status)
		i++
	}
	if f.RefundStatus != "" {
		where = append(where, "refund_status=$"+argN(i))
		args = append(args, f.RefundStatus)
		i++
	}

	whereClause := "1=1"
	if len(where) > 0 {
		whereClause = strings.Join(where, " AND ")
	}

	var total int64
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE "+whereClause, args...).Scan(&total)

	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, orderSelect+`
		WHERE `+whereClause+`
		ORDER BY created_at DESC
		LIMIT $`+argN(i)+` OFFSET $`+argN(i+1),
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		o, err := r.scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}

	totalPages := int(math.Ceil(float64(total) / float64(f.PageSize)))
	return &ports.Page[domain.Order]{
		Data: orders, Total: total, Page: f.Page, PageSize: f.PageSize,
		TotalPages: totalPages, HasNext: f.Page < totalPages, HasPrev: f.Page > 1,
	}, rows.Err()
}

func (r *OrderRepo) Create(ctx context.Context, o *domain.Order) (*domain.Order, error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	items, _ := json.Marshal(o.Items)
	addr, _ := json.Marshal(o.ShippingAddress)

	var coupon, notes *string
	if o.CouponCode != "" {
		coupon = &o.CouponCode
	}
	if o.Notes != "" {
		notes = &o.Notes
	}

	return r.scanOrder(r.db.QueryRow(ctx, `
		INSERT INTO orders (id, branch_id, customer_id, status, items, subtotal, tax, discount,
		                    shipping_cost, total, currency, shipping_address, coupon_code,
		                    reservation_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, branch_id, customer_id, status, items, subtotal, tax, discount,
		          shipping_cost, total, currency, shipping_address,
		          COALESCE(payment_intent_id,''), COALESCE(coupon_code,''),
		          COALESCE(reservation_id,''), COALESCE(notes,''),
		          COALESCE(refund_status,'none'), COALESCE(refund_reason,''), refunded_at,
		          created_at, updated_at`,
		o.ID, o.BranchID, o.CustomerID, o.Status, items, o.Subtotal, o.Tax, o.Discount,
		o.ShippingCost, o.Total, o.Currency, addr, coupon, o.ReservationID, notes))
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

func (r *OrderRepo) ConfirmPayment(ctx context.Context, orderID, paymentIntentID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET payment_intent_id=$2, status='processing', updated_at=NOW() WHERE id=$1`,
		orderID, paymentIntentID)
	return err
}

func (r *OrderRepo) RequestRefund(ctx context.Context, orderID, reason string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE orders
		SET refund_status='requested', refund_reason=$2, updated_at=NOW()
		WHERE id=$1 AND status='delivered' AND refund_status='none'`,
		orderID, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ports.ErrInvalidTransition
	}
	return nil
}

func (r *OrderRepo) ApproveRefund(ctx context.Context, orderID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE orders
		SET refund_status='approved', refunded_at=NOW(), status='refunded', updated_at=NOW()
		WHERE id=$1 AND refund_status='requested'`,
		orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ports.ErrInvalidTransition
	}
	return nil
}

func (r *OrderRepo) RejectRefund(ctx context.Context, orderID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE orders
		SET refund_status='rejected', updated_at=NOW()
		WHERE id=$1 AND refund_status='requested'`,
		orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ports.ErrInvalidTransition
	}
	return nil
}

func (r *OrderRepo) scanOrder(row rowScanner) (*domain.Order, error) {
	var o domain.Order
	var items, addr []byte
	if err := row.Scan(
		&o.ID, &o.BranchID, &o.CustomerID, &o.Status,
		&items, &o.Subtotal, &o.Tax, &o.Discount, &o.ShippingCost, &o.Total,
		&o.Currency, &addr,
		&o.PaymentIntentID, &o.CouponCode, &o.ReservationID, &o.Notes,
		&o.RefundStatus, &o.RefundReason, &o.RefundedAt,
		&o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	json.Unmarshal(items, &o.Items)
	json.Unmarshal(addr, &o.ShippingAddress)
	if o.Items == nil {
		o.Items = []domain.OrderItem{}
	}
	return &o, nil
}

func argN(n int) string {
	return strconv.Itoa(n)
}
