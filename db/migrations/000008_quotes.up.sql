CREATE TABLE IF NOT EXISTS quotes (
  id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_number  SERIAL,
  items         JSONB         NOT NULL DEFAULT '[]',
  subtotal      NUMERIC(14,2) NOT NULL DEFAULT 0,
  tax_rate      NUMERIC(5,4)  NOT NULL DEFAULT 0.07,
  tax_amount    NUMERIC(14,2) NOT NULL DEFAULT 0,
  total         NUMERIC(14,2) NOT NULL DEFAULT 0,
  currency      TEXT          NOT NULL DEFAULT 'USD',
  store_name    TEXT          NOT NULL DEFAULT '',
  contact_email TEXT          NOT NULL DEFAULT '',
  support_phone TEXT          NOT NULL DEFAULT '',
  customer_name  TEXT         NOT NULL DEFAULT '',
  customer_email TEXT         NOT NULL DEFAULT '',
  customer_phone TEXT         NOT NULL DEFAULT '',
  note          TEXT          NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMPTZ
);
