-- Seed de datos de prueba para ShopGo
-- Admin: admin@shopgo.com / Admin1234!
-- Customer: cliente@shopgo.com / Cliente123!

-- ─── Sucursal principal ───────────────────────────────────────────────────────
INSERT INTO branches (id, name, address, warehouse_mode, is_active)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'Tienda Principal',
  '{"street":"Av. Insurgentes 123","city":"Ciudad de México","state":"CDMX","zip":"06600","country":"MX"}',
  false,
  true
) ON CONFLICT DO NOTHING;

-- ─── Usuarios ─────────────────────────────────────────────────────────────────
-- Admin
INSERT INTO users (id, email, password_hash, role, branch_id, first_name, last_name, is_active)
VALUES (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'admin@shopgo.com',
  '$2a$10$UGfiVqUBd02ozAb2eqJY4OVUaEl5nB6/bPP3MvG2BjFXscjOR56MK',
  'admin',
  '11111111-1111-1111-1111-111111111111',
  'Admin',
  'ShopGo',
  true
) ON CONFLICT (email) DO NOTHING;

-- Customer
INSERT INTO users (id, email, password_hash, role, first_name, last_name, is_active)
VALUES (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'cliente@shopgo.com',
  '$2a$10$zrIJnfv7S5TZv2jBosuBSegtbuWRLmSPTbW5mevafugJ3Bghf.Y0K',
  'customer',
  'Juan',
  'Pérez',
  true
) ON CONFLICT (email) DO NOTHING;

-- ─── Categorías ───────────────────────────────────────────────────────────────
INSERT INTO categories (id, name, slug) VALUES
  ('cat-elect-0000-0000-000000000001', 'Electrónicos',   'electronicos'),
  ('cat-ropa-00000-0000-000000000002', 'Ropa',            'ropa'),
  ('cat-hogar-0000-0000-000000000003', 'Hogar',           'hogar'),
  ('cat-deport-000-0000-000000000004', 'Deportes',        'deportes')
ON CONFLICT DO NOTHING;

-- ─── Productos ────────────────────────────────────────────────────────────────
INSERT INTO products (id, sku, name, description, base_price, category_id, images, attributes, tags, is_active)
VALUES
(
  'prod-0001-0000-0000-000000000001',
  'ELEC-001',
  'Audífonos Bluetooth Pro',
  'Audífonos inalámbricos con cancelación de ruido activa, 30h de batería y sonido premium.',
  1299.00,
  'cat-elect-0000-0000-000000000001',
  '[{"url":"https://placehold.co/600x400?text=Audifonos","alt":"Audífonos Bluetooth Pro","is_primary":true}]',
  '{"color":"Negro","conectividad":"Bluetooth 5.0","bateria":"30h"}',
  ARRAY['audio','bluetooth','inalambrico'],
  true
),
(
  'prod-0002-0000-0000-000000000002',
  'ELEC-002',
  'Smartwatch Serie 5',
  'Reloj inteligente con monitor cardíaco, GPS integrado y resistencia al agua 50m.',
  2499.00,
  'cat-elect-0000-0000-000000000001',
  '[{"url":"https://placehold.co/600x400?text=Smartwatch","alt":"Smartwatch Serie 5","is_primary":true}]',
  '{"color":"Plateado","pantalla":"AMOLED 1.4\"","bateria":"7 días"}',
  ARRAY['smartwatch','gps','deportes'],
  true
),
(
  'prod-0003-0000-0000-000000000003',
  'ELEC-003',
  'Bocina Portátil 360°',
  'Bocina con sonido envolvente 360°, resistente al agua IPX7 y 12h de batería.',
  799.00,
  'cat-elect-0000-0000-000000000001',
  '[{"url":"https://placehold.co/600x400?text=Bocina","alt":"Bocina Portátil 360°","is_primary":true}]',
  '{"color":"Azul","watts":"20W","resistencia":"IPX7"}',
  ARRAY['audio','portatil','agua'],
  true
),
(
  'prod-0004-0000-0000-000000000004',
  'ROPA-001',
  'Playera Deportiva Dry-Fit',
  'Playera de alto rendimiento con tecnología de absorción de humedad, ideal para entrenamientos.',
  349.00,
  'cat-ropa-00000-0000-000000000002',
  '[{"url":"https://placehold.co/600x400?text=Playera","alt":"Playera Deportiva","is_primary":true}]',
  '{"talla":"M","material":"95% Poliéster","genero":"Unisex"}',
  ARRAY['ropa','deportes','dry-fit'],
  true
),
(
  'prod-0005-0000-0000-000000000005',
  'ROPA-002',
  'Sudadera Con Capucha',
  'Sudadera premium de algodón orgánico con forro polar interior. Perfecta para el clima frío.',
  699.00,
  'cat-ropa-00000-0000-000000000002',
  '[{"url":"https://placehold.co/600x400?text=Sudadera","alt":"Sudadera Con Capucha","is_primary":true}]',
  '{"talla":"L","material":"80% Algodón Orgánico","color":"Gris"}',
  ARRAY['ropa','invierno','casual'],
  true
),
(
  'prod-0006-0000-0000-000000000006',
  'HOGAR-001',
  'Cafetera Espresso Pro',
  'Cafetera con bomba de 15 bares de presión, vaporizador de leche y pantalla táctil.',
  3299.00,
  'cat-hogar-0000-0000-000000000003',
  '[{"url":"https://placehold.co/600x400?text=Cafetera","alt":"Cafetera Espresso","is_primary":true}]',
  '{"color":"Negro/Plata","presion":"15 bares","capacidad":"1.8L"}',
  ARRAY['cafe','cocina','premium'],
  true
),
(
  'prod-0007-0000-0000-000000000007',
  'HOGAR-002',
  'Set de Cuchillos Profesional',
  'Set de 5 cuchillos de acero inoxidable alemán con bloque de madera. Incluye afilador.',
  1599.00,
  'cat-hogar-0000-0000-000000000003',
  '[{"url":"https://placehold.co/600x400?text=Cuchillos","alt":"Set Cuchillos","is_primary":true}]',
  '{"material":"Acero Inox Alemán","piezas":6,"incluye":"Bloque + afilador"}',
  ARRAY['cocina','cuchillos','chef'],
  true
),
(
  'prod-0008-0000-0000-000000000008',
  'DEP-001',
  'Mochila Hiking 45L',
  'Mochila para senderismo con soporte lumbar ergonómico, cubierta impermeable y múltiples bolsillos.',
  1199.00,
  'cat-deport-000-0000-000000000004',
  '[{"url":"https://placehold.co/600x400?text=Mochila","alt":"Mochila Hiking","is_primary":true}]',
  '{"capacidad":"45L","color":"Verde/Negro","impermeable":true}',
  ARRAY['hiking','mochila','outdoor'],
  true
)
ON CONFLICT (sku) DO NOTHING;

-- ─── Inventario ───────────────────────────────────────────────────────────────
INSERT INTO inventory (product_id, branch_id, quantity, reserved_qty, reorder_point)
VALUES
  ('prod-0001-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 50,  0, 5),
  ('prod-0002-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 30,  0, 5),
  ('prod-0003-0000-0000-000000000003', '11111111-1111-1111-1111-111111111111', 75,  0, 10),
  ('prod-0004-0000-0000-000000000004', '11111111-1111-1111-1111-111111111111', 100, 0, 15),
  ('prod-0005-0000-0000-000000000005', '11111111-1111-1111-1111-111111111111', 60,  0, 10),
  ('prod-0006-0000-0000-000000000006', '11111111-1111-1111-1111-111111111111', 20,  0, 3),
  ('prod-0007-0000-0000-000000000007', '11111111-1111-1111-1111-111111111111', 35,  0, 5),
  ('prod-0008-0000-0000-000000000008', '11111111-1111-1111-1111-111111111111', 45,  0, 8)
ON CONFLICT DO NOTHING;

-- ─── Cupón de descuento ───────────────────────────────────────────────────────
INSERT INTO coupons (id, code, type, value, valid_from, valid_until, max_uses, is_active)
VALUES (
  'coupon-00-0000-0000-000000000001',
  'BIENVENIDO10',
  'percent',
  10.00,
  NOW(),
  NOW() + INTERVAL '1 year',
  1000,
  true
) ON CONFLICT DO NOTHING;

-- ─── Orden de ejemplo ─────────────────────────────────────────────────────────
INSERT INTO orders (id, branch_id, customer_id, status, items, subtotal, tax, total, currency, shipping_address)
VALUES (
  'order-001-0000-0000-000000000001',
  '11111111-1111-1111-1111-111111111111',
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'delivered',
  '[{"product_id":"prod-0001-0000-0000-000000000001","sku":"ELEC-001","name":"Audífonos Bluetooth Pro","qty":1,"unit_price":1299.00,"total":1299.00},{"product_id":"prod-0004-0000-0000-000000000004","sku":"ROPA-001","name":"Playera Deportiva Dry-Fit","qty":2,"unit_price":349.00,"total":698.00}]',
  1997.00,
  319.52,
  2316.52,
  'MXN',
  '{"street":"Calle Reforma 456","city":"Ciudad de México","state":"CDMX","zip":"06600","country":"MX"}'
) ON CONFLICT DO NOTHING;
