ALTER TABLE orders ADD COLUMN IF NOT EXISTS quote_id TEXT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_quote_id ON orders(quote_id) WHERE quote_id IS NOT NULL;
