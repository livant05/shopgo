-- name: GetRevenue :one
SELECT
  COALESCE(SUM(total), 0)            AS gmv,
  COUNT(*)                            AS orders,
  COALESCE(AVG(total), 0)            AS aov,
  COUNT(DISTINCT customer_id)         AS customers
FROM orders
WHERE status NOT IN ('pending','cancelled','refunded')
  AND created_at BETWEEN @from_date AND @to_date;

-- name: GetSalesByBranch :many
SELECT
  o.branch_id,
  b.name AS branch_name,
  COUNT(o.id) AS orders,
  COALESCE(SUM(o.total), 0) AS revenue,
  COUNT(DISTINCT o.customer_id) AS customers
FROM orders o
JOIN branches b ON b.id = o.branch_id
WHERE o.status NOT IN ('pending','cancelled','refunded')
  AND o.created_at BETWEEN @from_date AND @to_date
GROUP BY o.branch_id, b.name
ORDER BY revenue DESC;

-- name: GetTopProducts :many
SELECT
  p.id, p.name, p.sku,
  SUM((item->>'quantity')::int) AS units_sold,
  SUM((item->>'line_total')::numeric) AS revenue
FROM orders o
CROSS JOIN LATERAL jsonb_array_elements(o.items) AS item
JOIN products p ON p.id = (item->>'product_id')::uuid
WHERE o.status NOT IN ('pending','cancelled','refunded')
  AND o.created_at BETWEEN @from_date AND @to_date
  AND (@branch_id::uuid IS NULL OR o.branch_id = @branch_id)
GROUP BY p.id, p.name, p.sku
ORDER BY revenue DESC
LIMIT @top_n;

-- name: GetDailySeries :many
SELECT
  DATE_TRUNC('day', created_at)::date::text AS day,
  COUNT(*) AS orders,
  COALESCE(SUM(total), 0) AS revenue
FROM orders
WHERE status NOT IN ('pending','cancelled','refunded')
  AND created_at BETWEEN @from_date AND @to_date
  AND (@branch_id::uuid IS NULL OR branch_id = @branch_id)
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY day ASC;

-- name: GetInventoryReport :many
SELECT
  p.sku, p.name AS product_name, b.name AS branch_name,
  i.quantity, i.reserved_qty,
  (i.quantity - i.reserved_qty) AS available,
  i.reorder_point,
  ((i.quantity - i.reserved_qty) <= i.reorder_point) AS is_low,
  COALESCE(bp.price, p.base_price) AS price
FROM inventory i
JOIN products p ON p.id = i.product_id
JOIN branches b ON b.id = i.branch_id
LEFT JOIN branch_prices bp ON bp.product_id = i.product_id AND bp.branch_id = i.branch_id
WHERE (@branch_id::uuid IS NULL OR i.branch_id = @branch_id)
ORDER BY is_low DESC, p.name;
