package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/ports"
)

type ReportRepo struct{ db *pgxpool.Pool }

func NewReportRepo(db *pgxpool.Pool) *ReportRepo { return &ReportRepo{db} }

func (r *ReportRepo) Revenue(ctx context.Context, from, to string) (*ports.RevenueMetrics, error) {
	var m ports.RevenueMetrics
	err := r.db.QueryRow(ctx, `
		SELECT
		  COALESCE(SUM(total), 0),
		  COUNT(*),
		  COALESCE(AVG(total), 0),
		  COUNT(DISTINCT customer_id)
		FROM orders
		WHERE status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR created_at >= $1::timestamptz)
		  AND ($2 = '' OR created_at <= $2::timestamptz)`,
		from, to).
		Scan(&m.GMV, &m.Orders, &m.AOV, &m.Customers)
	if err != nil {
		return &ports.RevenueMetrics{}, nil
	}
	return &m, nil
}

func (r *ReportRepo) SalesByBranch(ctx context.Context, from, to string) ([]*ports.BranchSales, error) {
	rows, err := r.db.Query(ctx, `
		SELECT o.branch_id, b.name,
		       COUNT(*) AS orders,
		       COALESCE(SUM(o.total), 0) AS revenue,
		       COUNT(DISTINCT o.customer_id) AS customers
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE o.status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR o.created_at >= $1::timestamptz)
		  AND ($2 = '' OR o.created_at <= $2::timestamptz)
		GROUP BY o.branch_id, b.name
		ORDER BY revenue DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.BranchSales, 0)
	for rows.Next() {
		var s ports.BranchSales
		if err := rows.Scan(&s.BranchID, &s.BranchName, &s.Orders, &s.Revenue, &s.Customers); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *ReportRepo) TopProducts(ctx context.Context, branchID, from, to string, n int) ([]*ports.TopProduct, error) {
	if n <= 0 {
		n = 10
	}
	rows, err := r.db.Query(ctx, `
		SELECT
		  item->>'product_id' AS product_id,
		  item->>'name'       AS name,
		  item->>'sku'        AS sku,
		  SUM((item->>'quantity')::int) AS units_sold,
		  SUM((item->>'line_total')::numeric) AS revenue
		FROM orders o,
		     jsonb_array_elements(o.items) AS item
		WHERE o.status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR o.branch_id = $1)
		  AND ($2 = '' OR o.created_at >= $2::timestamptz)
		  AND ($3 = '' OR o.created_at <= $3::timestamptz)
		GROUP BY product_id, name, sku
		ORDER BY units_sold DESC
		LIMIT $4`, branchID, from, to, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.TopProduct, 0)
	for rows.Next() {
		var p ports.TopProduct
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.UnitsSold, &p.Revenue); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *ReportRepo) TopCustomers(ctx context.Context, from, to string, n int) ([]*ports.TopCustomer, error) {
	if n <= 0 {
		n = 10
	}
	rows, err := r.db.Query(ctx, `
		SELECT o.customer_id,
		       COALESCE(u.email, ''),
		       COALESCE(u.first_name || ' ' || u.last_name, ''),
		       COUNT(*) AS orders,
		       COALESCE(SUM(o.total), 0) AS revenue
		FROM orders o
		LEFT JOIN users u ON u.id = o.customer_id
		WHERE o.status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR o.created_at >= $1::timestamptz)
		  AND ($2 = '' OR o.created_at <= $2::timestamptz)
		GROUP BY o.customer_id, u.email, u.first_name, u.last_name
		ORDER BY revenue DESC
		LIMIT $3`, from, to, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.TopCustomer, 0)
	for rows.Next() {
		var c ports.TopCustomer
		if err := rows.Scan(&c.CustomerID, &c.Email, &c.FullName, &c.Orders, &c.Revenue); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (r *ReportRepo) HourlySeries(ctx context.Context, branchID, from, to string) ([]*ports.HourlyStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT EXTRACT(HOUR FROM created_at)::int AS hour,
		       COUNT(*) AS orders,
		       COALESCE(SUM(total), 0) AS revenue
		FROM orders
		WHERE status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR branch_id = $1)
		  AND ($2 = '' OR created_at >= $2::timestamptz)
		  AND ($3 = '' OR created_at <= $3::timestamptz)
		GROUP BY hour
		ORDER BY hour ASC`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.HourlyStat, 0)
	for rows.Next() {
		var s ports.HourlyStat
		if err := rows.Scan(&s.Hour, &s.Orders, &s.Revenue); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *ReportRepo) DailySeries(ctx context.Context, branchID, from, to string) ([]*ports.DailyStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
		  DATE(created_at)::text AS day,
		  COUNT(*) AS orders,
		  COALESCE(SUM(total), 0) AS revenue
		FROM orders
		WHERE status NOT IN ('cancelled','refunded')
		  AND ($1 = '' OR branch_id = $1)
		  AND ($2 = '' OR created_at >= $2::timestamptz)
		  AND ($3 = '' OR created_at <= $3::timestamptz)
		GROUP BY DATE(created_at)
		ORDER BY day ASC`, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.DailyStat, 0)
	for rows.Next() {
		var s ports.DailyStat
		if err := rows.Scan(&s.Day, &s.Orders, &s.Revenue); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *ReportRepo) InventoryReport(ctx context.Context, branchID string) ([]*ports.InvRow, error) {
	q := `
		SELECT p.sku, p.name, b.name,
		       i.quantity, i.reserved_qty,
		       i.quantity - i.reserved_qty,
		       i.reorder_point,
		       (i.quantity - i.reserved_qty) <= i.reorder_point,
		       COALESCE(bp.price, p.base_price)
		FROM inventory i
		JOIN products p  ON p.id = i.product_id
		JOIN branches b  ON b.id = i.branch_id
		LEFT JOIN branch_prices bp ON bp.product_id = i.product_id AND bp.branch_id = i.branch_id
		WHERE p.is_active = true`
	args := []any{}
	if branchID != "" {
		q += " AND i.branch_id=$1"
		args = append(args, branchID)
	}
	q += " ORDER BY p.name, b.name"

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*ports.InvRow, 0)
	for rows.Next() {
		var row ports.InvRow
		if err := rows.Scan(&row.SKU, &row.ProductName, &row.BranchName,
			&row.Quantity, &row.ReservedQty, &row.Available,
			&row.ReorderPoint, &row.IsLow, &row.Price); err != nil {
			return nil, err
		}
		list = append(list, &row)
	}
	return list, rows.Err()
}
