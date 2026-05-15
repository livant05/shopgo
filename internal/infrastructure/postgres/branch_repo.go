package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type BranchRepo struct{ db *pgxpool.Pool }

func NewBranchRepo(db *pgxpool.Pool) *BranchRepo { return &BranchRepo{db} }

func (r *BranchRepo) GetByID(ctx context.Context, id string) (*domain.Branch, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT id, name, address, warehouse_mode, settings, is_active, created_at
		 FROM branches WHERE id = $1`, id))
}

func (r *BranchRepo) List(ctx context.Context) ([]*domain.Branch, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, address, warehouse_mode, settings, is_active, created_at
		 FROM branches ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*domain.Branch, 0)
	for rows.Next() {
		b, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (r *BranchRepo) Create(ctx context.Context, b *domain.Branch) (*domain.Branch, error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	addr, _ := json.Marshal(b.Address)
	settings, _ := json.Marshal(b.Settings)
	return r.scan(r.db.QueryRow(ctx,
		`INSERT INTO branches (id, name, address, warehouse_mode, settings, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, name, address, warehouse_mode, settings, is_active, created_at`,
		b.ID, b.Name, addr, b.WarehouseMode, settings, true))
}

func (r *BranchRepo) Update(ctx context.Context, b *domain.Branch) (*domain.Branch, error) {
	addr, _ := json.Marshal(b.Address)
	settings, _ := json.Marshal(b.Settings)
	_, err := r.db.Exec(ctx,
		`UPDATE branches SET name=$2, address=$3, warehouse_mode=$4, settings=$5 WHERE id=$1`,
		b.ID, b.Name, addr, b.WarehouseMode, settings)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BranchRepo) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.Exec(ctx, `UPDATE branches SET is_active=$2 WHERE id=$1`, id, active)
	return err
}

type rowScanner interface{ Scan(dest ...any) error }

func (r *BranchRepo) scan(row rowScanner) (*domain.Branch, error) {
	var b domain.Branch
	var addr, settings []byte
	if err := row.Scan(&b.ID, &b.Name, &addr, &b.WarehouseMode, &settings, &b.IsActive, &b.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	json.Unmarshal(addr, &b.Address)
	json.Unmarshal(settings, &b.Settings)
	return &b, nil
}
