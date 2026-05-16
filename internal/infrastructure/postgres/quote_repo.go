package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

func (r *QuoteRepo) List(ctx context.Context, f ports.QuoteFilter) (*ports.Page[domain.Quote], error) {
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PageSize

	where := []string{"1=1"}
	args := []any{}
	n := 1

	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		where = append(where, fmt.Sprintf(
			"(LOWER(customer_name) LIKE $%d OR LOWER(customer_email) LIKE $%d)", n, n+1))
		args = append(args, like, like)
		n += 2
	}
	if f.From != "" {
		where = append(where, fmt.Sprintf("created_at >= $%d", n))
		args = append(args, f.From)
		n++
	}
	if f.To != "" {
		where = append(where, fmt.Sprintf("created_at <= $%d", n))
		args = append(args, f.To)
		n++
	}

	clause := strings.Join(where, " AND ")

	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM quotes WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, err
	}

	limitArgs := append(args, f.PageSize, offset)
	rows, err := r.db.Query(ctx,
		`SELECT `+quoteSelect+` FROM quotes WHERE `+clause+
			fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n, n+1),
		limitArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []domain.Quote
	for rows.Next() {
		q, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, *q)
	}
	if quotes == nil {
		quotes = []domain.Quote{}
	}

	pageSize := f.PageSize
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &ports.Page[domain.Quote]{
		Data:       quotes,
		Total:      total,
		Page:       f.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    f.Page < totalPages,
		HasPrev:    f.Page > 1,
	}, nil
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
