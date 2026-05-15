-- name: GetProduct :one
SELECT * FROM products WHERE id = @id;

-- name: GetProductWithPrice :one
SELECT p.*, COALESCE(bp.price, p.base_price) AS effective_price,
  (SELECT quantity - reserved_qty FROM inventory WHERE product_id = p.id AND branch_id = @branch_id) AS stock
FROM products p
LEFT JOIN branch_prices bp ON bp.product_id = p.id AND bp.branch_id = @branch_id
WHERE p.id = @product_id AND p.is_active = true;

-- name: ListProducts :many
SELECT p.*, COALESCE(bp.price, p.base_price) AS effective_price
FROM products p
LEFT JOIN branch_prices bp ON bp.product_id = p.id AND bp.branch_id = @branch_id
WHERE
  p.is_active = true
  AND (@search::text IS NULL OR p.search_vector @@ plainto_tsquery('spanish', @search))
  AND (@category_id::uuid IS NULL OR p.category_id = @category_id)
ORDER BY
  CASE WHEN @sort = 'price_asc'  THEN COALESCE(bp.price, p.base_price) END ASC,
  CASE WHEN @sort = 'price_desc' THEN COALESCE(bp.price, p.base_price) END DESC,
  CASE WHEN @sort = 'name'       THEN p.name END ASC,
  p.created_at DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountProducts :one
SELECT COUNT(*) FROM products p
WHERE
  p.is_active = true
  AND (@search::text IS NULL OR p.search_vector @@ plainto_tsquery('spanish', @search))
  AND (@category_id::uuid IS NULL OR p.category_id = @category_id);

-- name: CreateProduct :one
INSERT INTO products (sku, name, description, base_price, category_id, images, attributes, tags)
VALUES (@sku, @name, @description, @base_price, @category_id, @images, @attributes, @tags)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products SET
  name = @name, description = @description, base_price = @base_price,
  category_id = @category_id, images = @images, attributes = @attributes,
  tags = @tags, updated_at = NOW()
WHERE id = @id RETURNING *;

-- name: SetProductActive :exec
UPDATE products SET is_active = @is_active, updated_at = NOW() WHERE id = @id;

-- name: SetBranchPrice :exec
INSERT INTO branch_prices (product_id, branch_id, price)
VALUES (@product_id, @branch_id, @price)
ON CONFLICT (product_id, branch_id) DO UPDATE SET price = @price, updated_at = NOW();

-- name: ListCategories :many
SELECT * FROM categories ORDER BY name;
