-- 002_indexes.sql — índices para performance de producción

-- Búsqueda full-text en productos (GIN)
CREATE INDEX IF NOT EXISTS idx_products_search
  ON products USING gin(search_vector);

-- Productos activos por categoría (storefront)
CREATE INDEX IF NOT EXISTS idx_products_active_cat
  ON products(category_id, created_at DESC) WHERE is_active = true;

-- Inventario con INCLUDE (evita heap fetch en reservas de stock)
CREATE INDEX IF NOT EXISTS idx_inventory_branch_product
  ON inventory(branch_id, product_id) INCLUDE (quantity, reserved_qty);

-- Órdenes activas por sucursal y fecha (query más frecuente del panel)
CREATE INDEX IF NOT EXISTS idx_orders_branch_date
  ON orders(branch_id, created_at DESC)
  WHERE status NOT IN ('cancelled','refunded');

-- Órdenes por cliente (historial de compras)
CREATE INDEX IF NOT EXISTS idx_orders_customer
  ON orders(customer_id, created_at DESC);

-- Órdenes por estado (queue del panel)
CREATE INDEX IF NOT EXISTS idx_orders_status
  ON orders(status, created_at DESC);

-- Usuarios activos por rol y sucursal
CREATE INDEX IF NOT EXISTS idx_users_role_branch
  ON users(role, branch_id) WHERE is_active = true;

-- Movimientos de inventario por producto y fecha
CREATE INDEX IF NOT EXISTS idx_inv_movements_product
  ON inventory_movements(product_id, created_at DESC);

-- Cupones activos por código (validación en checkout)
CREATE INDEX IF NOT EXISTS idx_coupons_code
  ON coupons(code) WHERE is_active = true;

-- Audit log por tabla/registro
CREATE INDEX IF NOT EXISTS idx_audit_table_record
  ON audit_log(table_name, record_id, created_at DESC);

-- Limpieza automática de tokens expirados (pg_cron recomendado)
CREATE INDEX IF NOT EXISTS idx_revoked_tokens_expiry
  ON revoked_tokens(expires_at);
