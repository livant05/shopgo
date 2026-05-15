package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
)

type StoreRepo struct{ db *pgxpool.Pool }

func NewStoreRepo(db *pgxpool.Pool) *StoreRepo { return &StoreRepo{db} }

func (r *StoreRepo) Get(ctx context.Context) (*domain.StoreConfig, error) {
	var sc domain.StoreConfig
	err := r.db.QueryRow(ctx, `
		SELECT store_name, COALESCE(logo_url,''), currency, tax_rate, tax_inclusive,
		       contact_email, COALESCE(support_phone,'')
		FROM store_config LIMIT 1`).
		Scan(&sc.StoreName, &sc.LogoURL, &sc.Currency, &sc.TaxRate, &sc.TaxInclusive,
			&sc.ContactEmail, &sc.SupportPhone)
	if err != nil {
		return &domain.StoreConfig{StoreName: "Mi Tienda", Currency: "MXN", TaxRate: 0.16}, nil
	}
	return &sc, nil
}

func (r *StoreRepo) Update(ctx context.Context, sc *domain.StoreConfig) error {
	_, err := r.db.Exec(ctx, `
		UPDATE store_config
		SET store_name=$1, logo_url=$2, currency=$3, tax_rate=$4,
		    tax_inclusive=$5, contact_email=$6, support_phone=$7, updated_at=NOW()`,
		sc.StoreName, sc.LogoURL, sc.Currency, sc.TaxRate,
		sc.TaxInclusive, sc.ContactEmail, sc.SupportPhone)
	return err
}

// CouponRepo — acceso directo a cupones para validación en checkout.
type CouponRepo struct{ db *pgxpool.Pool }

func NewCouponRepo(db *pgxpool.Pool) *CouponRepo { return &CouponRepo{db} }

func (r *CouponRepo) Validate(ctx context.Context, code string, subtotal float64) (*domain.Coupon, float64, error) {
	var c domain.Coupon
	err := r.db.QueryRow(ctx, `
		SELECT id, code, type, value, valid_until, max_uses, uses_count, is_active
		FROM coupons
		WHERE code = $1
		  AND is_active = true
		  AND valid_from <= NOW()
		  AND (valid_until IS NULL OR valid_until >= NOW())
		  AND (max_uses IS NULL OR uses_count < max_uses)`, code).
		Scan(&c.ID, &c.Code, &c.Type, &c.Value, &c.ValidUntil, &c.MaxUses, &c.UsesCount, &c.IsActive)
	if err != nil {
		return nil, 0, err
	}
	discount := c.Apply(subtotal)
	return &c, discount, nil
}
