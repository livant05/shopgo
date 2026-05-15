-- Agrega campos extra a store_config para integraciones y tema.
ALTER TABLE store_config
  ADD COLUMN IF NOT EXISTS stripe_public_key TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS social_instagram   TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS social_facebook    TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS social_whatsapp    TEXT NOT NULL DEFAULT '';
