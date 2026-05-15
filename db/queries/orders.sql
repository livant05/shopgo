-- name: GetOrder :one
SELECT o.*, b.name AS branch_name FROM orders o
JOIN branches b ON b.id = o.branch_id
WHERE o.id = @id;

-- name: ListOrders :many
SELECT o.* FROM orders o
WHERE
  (@branch_id::uuid IS NULL OR o.branch_id = @branch_id)
  AND (@status::order_status IS NULL OR o.status = @status)
  AND o.created_at BETWEEN @from_date AND @to_date
ORDER BY o.created_at DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountOrders :one
SELECT COUNT(*) FROM orders
WHERE
  (@branch_id::uuid IS NULL OR branch_id = @branch_id)
  AND (@status::order_status IS NULL OR status = @status)
  AND created_at BETWEEN @from_date AND @to_date;

-- name: CreateOrder :one
INSERT INTO orders (branch_id, customer_id, status, items, subtotal, tax, discount,
  shipping_cost, total, currency, shipping_address, coupon_code, reservation_id, notes)
VALUES (@branch_id, @customer_id, 'pending', @items, @subtotal, @tax, @discount,
  @shipping_cost, @total, @currency, @shipping_address, @coupon_code, @reservation_id, @notes)
RETURNING *;

-- name: UpdateOrderStatus :exec
UPDATE orders SET status = @status, updated_at = NOW() WHERE id = @id;

-- name: ConfirmPayment :exec
UPDATE orders SET
  status = 'confirmed',
  payment_intent_id = @payment_intent_id,
  updated_at = NOW()
WHERE id = @id AND status = 'pending';

-- name: GetCashClose :many
SELECT
  DATE(created_at) AS date,
  COUNT(*) AS order_count,
  SUM(total) AS revenue,
  SUM(tax) AS tax_collected
FROM orders
WHERE branch_id = @branch_id
  AND DATE(created_at) = @date
  AND status NOT IN ('pending','cancelled','refunded')
GROUP BY DATE(created_at);
