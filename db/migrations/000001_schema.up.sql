-- 001_schema.sql
-- Schema único — sin complejidad de multi-tenant.
-- Una empresa, una BD, un schema.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TYPE user_role AS ENUM ('admin','manager','staff','customer');
CREATE TYPE order_status AS ENUM ('pending','confirmed','processing','shipped','delivered','cancelled','refunded');

-- Configuración global de la tienda (1 sola fila)
CREATE TABLE store_config (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  store_name   TEXT NOT NULL DEFAULT 'Mi Tienda',
  logo_url     TEXT,
  currency     TEXT NOT NULL DEFAULT 'MXN',
  tax_rate     NUMERIC(5,4) NOT NULL DEFAULT 0.16,
  tax_inclusive BOOLEAN NOT NULL DEFAULT false,
  contact_email TEXT NOT NULL DEFAULT '',
  support_phone TEXT,
  theme        TEXT NOT NULL DEFAULT 'light',
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO store_config DEFAULT VALUES;

-- Sucursales
CREATE TABLE branches (
  id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name           TEXT NOT NULL,
  address        JSONB NOT NULL DEFAULT '{}',
  warehouse_mode BOOLEAN NOT NULL DEFAULT false,
  settings       JSONB NOT NULL DEFAULT '{"currency":"MXN","tax_rate":0.16}',
  is_active      BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Usuarios del sistema (staff + admin) Y clientes
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email         CITEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role          user_role NOT NULL DEFAULT 'customer',
  branch_id     UUID REFERENCES branches(id) ON DELETE SET NULL,
  first_name    TEXT NOT NULL DEFAULT '',
  last_name     TEXT NOT NULL DEFAULT '',
  phone         TEXT,
  mfa_secret    TEXT,
  mfa_enabled   BOOLEAN NOT NULL DEFAULT false,
  is_active     BOOLEAN NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Categorías
CREATE TABLE categories (
  id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name      TEXT NOT NULL,
  slug      TEXT NOT NULL UNIQUE,
  parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  path      TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Productos
CREATE TABLE products (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  sku           TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  base_price    NUMERIC(12,4) NOT NULL CHECK (base_price >= 0),
  category_id   UUID REFERENCES categories(id) ON DELETE SET NULL,
  images        JSONB NOT NULL DEFAULT '[]',
  attributes    JSONB NOT NULL DEFAULT '{}',
  tags          TEXT[] NOT NULL DEFAULT '{}',
  is_active     BOOLEAN NOT NULL DEFAULT true,
  search_vector TSVECTOR GENERATED ALWAYS AS (
    setweight(to_tsvector('spanish', name), 'A') ||
    setweight(to_tsvector('spanish', COALESCE(description, '')), 'B')
  ) STORED,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Precios diferenciados por sucursal (opcional)
CREATE TABLE branch_prices (
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  branch_id  UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
  price      NUMERIC(12,4) NOT NULL CHECK (price >= 0),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (product_id, branch_id)
);

-- Inventario
CREATE TABLE inventory (
  product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  branch_id     UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
  quantity      INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  reserved_qty  INT NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
  reorder_point INT NOT NULL DEFAULT 5,
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (product_id, branch_id),
  CHECK (reserved_qty <= quantity)
);

-- Movimientos de inventario (historial completo)
CREATE TABLE inventory_movements (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  product_id    UUID NOT NULL REFERENCES products(id),
  from_branch_id UUID REFERENCES branches(id),
  to_branch_id   UUID REFERENCES branches(id),
  quantity      INT NOT NULL,
  type          TEXT NOT NULL, -- 'adjustment', 'transfer', 'sale', 'return'
  reason        TEXT NOT NULL,
  note          TEXT,
  user_id       UUID NOT NULL REFERENCES users(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Cupones de descuento
CREATE TABLE coupons (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  code        TEXT NOT NULL UNIQUE,
  type        TEXT NOT NULL CHECK (type IN ('percent','fixed')),
  value       NUMERIC(10,4) NOT NULL CHECK (value > 0),
  valid_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  valid_until TIMESTAMPTZ,
  max_uses    INT,
  uses_count  INT NOT NULL DEFAULT 0,
  is_active   BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Órdenes
CREATE TABLE orders (
  id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  branch_id        UUID NOT NULL REFERENCES branches(id),
  customer_id      UUID NOT NULL REFERENCES users(id),
  status           order_status NOT NULL DEFAULT 'pending',
  items            JSONB NOT NULL DEFAULT '[]',
  subtotal         NUMERIC(12,4) NOT NULL,
  tax              NUMERIC(12,4) NOT NULL DEFAULT 0,
  discount         NUMERIC(12,4) NOT NULL DEFAULT 0,
  shipping_cost    NUMERIC(12,4) NOT NULL DEFAULT 0,
  total            NUMERIC(12,4) NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'MXN',
  shipping_address JSONB NOT NULL DEFAULT '{}',
  payment_intent_id TEXT,
  coupon_code      TEXT,
  reservation_id   TEXT,
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Log de auditoría (inmutable)
CREATE TABLE audit_log (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  table_name  TEXT NOT NULL,
  record_id   UUID,
  action      TEXT NOT NULL,
  old_values  JSONB,
  new_values  JSONB,
  user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
REVOKE UPDATE, DELETE ON audit_log FROM PUBLIC;

-- Refresh tokens revocados (blacklist)
CREATE TABLE revoked_tokens (
  jti        TEXT PRIMARY KEY,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
