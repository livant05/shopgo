-- 003_triggers.sql — triggers de auditoría

CREATE OR REPLACE FUNCTION audit_fn()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO audit_log(table_name, record_id, action, old_values, new_values)
  VALUES (
    TG_TABLE_NAME,
    COALESCE(NEW.id, OLD.id),
    TG_OP,
    CASE WHEN TG_OP != 'INSERT' THEN row_to_json(OLD)::jsonb END,
    CASE WHEN TG_OP != 'DELETE' THEN row_to_json(NEW)::jsonb END
  );
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Auditar cambios de precio en productos
CREATE TRIGGER audit_products
  AFTER UPDATE ON products
  FOR EACH ROW
  WHEN (OLD.base_price IS DISTINCT FROM NEW.base_price OR OLD.is_active IS DISTINCT FROM NEW.is_active)
  EXECUTE FUNCTION audit_fn();

-- Auditar todos los cambios de órdenes
CREATE TRIGGER audit_orders
  AFTER INSERT OR UPDATE ON orders
  FOR EACH ROW EXECUTE FUNCTION audit_fn();

-- Auditar cambios críticos de usuarios
CREATE OR REPLACE FUNCTION audit_users_fn()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'UPDATE' AND (
    OLD.role IS DISTINCT FROM NEW.role OR
    OLD.is_active IS DISTINCT FROM NEW.is_active OR
    OLD.branch_id IS DISTINCT FROM NEW.branch_id
  ) THEN
    INSERT INTO audit_log(table_name, record_id, action, old_values, new_values)
    VALUES ('users', NEW.id, 'SENSITIVE_UPDATE',
      jsonb_build_object('role', OLD.role, 'is_active', OLD.is_active),
      jsonb_build_object('role', NEW.role, 'is_active', NEW.is_active));
  END IF;
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_users
  AFTER UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION audit_users_fn();
