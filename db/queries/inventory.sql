-- name: GetInventory :one
SELECT i.*, p.name AS product_name, p.sku, b.name AS branch_name
FROM inventory i
JOIN products p ON p.id = i.product_id
JOIN branches b ON b.id = i.branch_id
WHERE i.product_id = @product_id AND i.branch_id = @branch_id;

-- name: ListInventory :many
SELECT i.*, p.name AS product_name, p.sku, b.name AS branch_name,
  (i.quantity - i.reserved_qty) AS available
FROM inventory i
JOIN products p ON p.id = i.product_id
JOIN branches b ON b.id = i.branch_id
WHERE (@branch_id::uuid IS NULL OR i.branch_id = @branch_id)
ORDER BY p.name;

-- name: GetLowStock :many
SELECT i.*, p.name AS product_name, p.sku, b.name AS branch_name
FROM inventory i
JOIN products p ON p.id = i.product_id
JOIN branches b ON b.id = i.branch_id
WHERE
  (@branch_id::uuid IS NULL OR i.branch_id = @branch_id)
  AND (i.quantity - i.reserved_qty) <= i.reorder_point
ORDER BY (i.quantity - i.reserved_qty) ASC;

-- name: ReserveStock :one
-- Reserva atómica — sin deadlocks, sin transacción explícita
UPDATE inventory
SET reserved_qty = reserved_qty + @qty, updated_at = NOW()
WHERE product_id = @product_id
  AND branch_id = @branch_id
  AND (quantity - reserved_qty) >= @qty
RETURNING *;

-- name: CommitReservation :exec
UPDATE inventory
SET quantity = quantity - @qty, reserved_qty = reserved_qty - @qty, updated_at = NOW()
WHERE product_id = @product_id AND branch_id = @branch_id;

-- name: ReleaseReservation :exec
UPDATE inventory
SET reserved_qty = GREATEST(0, reserved_qty - @qty), updated_at = NOW()
WHERE product_id = @product_id AND branch_id = @branch_id;

-- name: UpsertInventory :one
INSERT INTO inventory (product_id, branch_id, quantity, reorder_point)
VALUES (@product_id, @branch_id, @quantity, @reorder_point)
ON CONFLICT (product_id, branch_id) DO UPDATE
SET quantity = inventory.quantity + @delta, updated_at = NOW()
RETURNING *;

-- name: TransferStock :exec
WITH deduct AS (
  UPDATE inventory
  SET quantity = quantity - @qty, updated_at = NOW()
  WHERE product_id = @product_id AND branch_id = @from_branch
    AND (quantity - reserved_qty) >= @qty
  RETURNING 1
)
INSERT INTO inventory (product_id, branch_id, quantity)
SELECT @product_id, @to_branch, @qty
WHERE EXISTS (SELECT 1 FROM deduct)
ON CONFLICT (product_id, branch_id) DO UPDATE
SET quantity = inventory.quantity + @qty, updated_at = NOW();
