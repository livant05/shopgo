// cmd/seed/main.go — creates an admin user and sample data. Run once.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://shopuser:shoppass@localhost:5432/shopgo?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("conectar BD:", err)
	}
	defer pool.Close()

	// ── Verificar tablas ─────────────────────────────────
	var tableCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public'`).Scan(&tableCount)
	fmt.Printf("Tablas en BD: %d\n", tableCount)

	if tableCount == 0 {
		log.Fatal("La BD está vacía. Corre las migraciones primero:\n  ~/go/bin/migrate -path db/migrations -database $DATABASE_URL up")
	}

	// ── Sucursal principal ───────────────────────────────
	branchID := "00000000-0000-0000-0000-000000000001"
	var existingBranch string
	pool.QueryRow(ctx, `SELECT id FROM branches WHERE id=$1`, branchID).Scan(&existingBranch)
	if existingBranch == "" {
		_, err = pool.Exec(ctx, `
			INSERT INTO branches (id, name, address, warehouse_mode, settings, is_active)
			VALUES ($1, 'Sucursal Principal', '{"street":"Av. Principal 1","city":"CDMX","state":"CDMX","zip":"06600","country":"MX"}'::jsonb,
			        false, '{"tax_rate":0.16,"currency":"MXN"}'::jsonb, true)
			ON CONFLICT DO NOTHING`,
			branchID)
		if err != nil {
			log.Fatal("crear sucursal:", err)
		}
		fmt.Println("✓ Sucursal principal creada")
	} else {
		fmt.Println("✓ Sucursal ya existe")
	}

	// ── Usuario admin ────────────────────────────────────
	email := "admin@tienda.com"
	password := "Admin1234!"

	var existingUser string
	pool.QueryRow(ctx, `SELECT id FROM users WHERE email=$1`, email).Scan(&existingUser)
	if existingUser != "" {
		fmt.Printf("✓ Usuario admin ya existe: %s\n", email)
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			log.Fatal("hash password:", err)
		}
		userID := uuid.New().String()
		_, err = pool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, role, branch_id, first_name, last_name, is_active)
			VALUES ($1, $2, $3, 'admin', $4, 'Admin', 'ShopGo', true)`,
			userID, email, string(hash), branchID)
		if err != nil {
			log.Fatal("crear usuario admin:", err)
		}
		fmt.Printf("✓ Usuario admin creado\n  Email:    %s\n  Password: %s\n", email, password)
	}

	// ── StoreConfig ──────────────────────────────────────
	var storeCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM store_config`).Scan(&storeCount)
	if storeCount == 0 {
		_, err = pool.Exec(ctx, `
			INSERT INTO store_config (store_name, currency, tax_rate, tax_inclusive, contact_email)
			VALUES ('Mi Tienda ShopGo', 'MXN', 0.16, false, 'admin@tienda.com')`)
		if err != nil {
			log.Printf("store_config: %v (puede no existir la tabla aún)", err)
		} else {
			fmt.Println("✓ Configuración de tienda creada")
		}
	}

	// ── Categoría y producto de ejemplo ─────────────────
	catID := uuid.New().String()
	var existingCat string
	pool.QueryRow(ctx, `SELECT id FROM categories WHERE slug='general'`).Scan(&existingCat)
	if existingCat == "" {
		_, err = pool.Exec(ctx, `INSERT INTO categories (id, name, slug) VALUES ($1,'General','general')`, catID)
		if err != nil {
			log.Fatal("crear categoría:", err)
		}
		fmt.Println("✓ Categoría 'General' creada")
	} else {
		catID = existingCat
		fmt.Println("✓ Categoría ya existe")
	}

	var productCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&productCount)
	if productCount == 0 {
		prodID := uuid.New().String()
		_, err = pool.Exec(ctx, `
			INSERT INTO products (id, sku, name, description, category_id, base_price, images, tags, attributes, is_active)
			VALUES ($1, 'PROD-001', 'Producto Demo', 'Producto de ejemplo para ShopGo', $2,
			        199.99, '[]'::jsonb, ARRAY['demo','ejemplo'], '{}'::jsonb, true)`,
			prodID, catID)
		if err != nil {
			log.Printf("crear producto: %v", err)
		} else {
			// Agregar al inventario de la sucursal
			pool.Exec(ctx, `
				INSERT INTO inventory (product_id, branch_id, quantity, reserved_qty, reorder_point)
				VALUES ($1, $2, 100, 0, 10) ON CONFLICT DO NOTHING`,
				prodID, branchID)
			fmt.Println("✓ Producto demo creado con 100 unidades en inventario")
		}
	} else {
		fmt.Printf("✓ Ya hay %d producto(s) en la BD\n", productCount)
	}

	fmt.Println("\n¡Listo! Puedes iniciar el servidor con:")
	fmt.Println("  set -a && source .env && set +a && ~/go/bin/air")
}
