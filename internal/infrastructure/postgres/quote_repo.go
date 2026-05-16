package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type QuoteRepo struct{ db *pgxpool.Pool }

func NewQuoteRepo(db *pgxpool.Pool) *QuoteRepo { return &QuoteRepo{db} }

const quoteSelect = `id, quote_number, items, subtotal, tax_rate, tax_amount, total, currency,
	store_name, contact_email, support_phone,
	customer_name, customer_email, customer_phone, note, created_at, expires_at`

func (r *QuoteRepo) Create(ctx context.Context, q *domain.Quote) (*domain.Quote, error) {
	items, _ := json.Marshal(q.Items)
	return r.scan(r.db.QueryRow(ctx, `
		INSERT INTO quotes
		  (items, subtotal, tax_rate, tax_amount, total, currency,
		   store_name, contact_email, support_phone,
		   customer_name, customer_email, customer_phone, note, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+quoteSelect,
		items, q.Subtotal, q.TaxRate, q.TaxAmount, q.Total, q.Currency,
		q.StoreName, q.ContactEmail, q.SupportPhone,
		q.CustomerName, q.CustomerEmail, q.CustomerPhone, q.Note, q.ExpiresAt))
}

func (r *QuoteRepo) GetByID(ctx context.Context, id string) (*domain.Quote, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+quoteSelect+` FROM quotes WHERE id = $1`, id))
}

func (r *QuoteRepo) scan(row rowScanner) (*domain.Quote, error) {
	var q domain.Quote
	var items []byte
	if err := row.Scan(
		&q.ID, &q.QuoteNumber, &items,
		&q.Subtotal, &q.TaxRate, &q.TaxAmount, &q.Total, &q.Currency,
		&q.StoreName, &q.ContactEmail, &q.SupportPhone,
		&q.CustomerName, &q.CustomerEmail, &q.CustomerPhone, &q.Note,
		&q.CreatedAt, &q.ExpiresAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	json.Unmarshal(items, &q.Items)
	if q.Items == nil {
		q.Items = []domain.QuoteItem{}
	}
	return &q, nil
}
