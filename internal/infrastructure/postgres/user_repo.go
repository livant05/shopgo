package postgres

import (
	"context"
	"encoding/json"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db} }

const userCols = `id, email, password_hash, role, COALESCE(branch_id::text,''),
	first_name, last_name, COALESCE(phone,''), COALESCE(default_address,'{}'),
	COALESCE(mfa_secret,''), mfa_enabled, is_active, created_at, updated_at`

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE email = $1`, email))
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Role == "" {
		u.Role = domain.RoleCustomer
	}
	var branchID *string
	if u.BranchID != "" {
		branchID = &u.BranchID
	}
	return r.scan(r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash, role, branch_id, first_name, last_name, phone, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING `+userCols,
		u.ID, u.Email, u.PasswordHash, u.Role, branchID,
		u.FirstName, u.LastName, u.Phone, true))
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	var branchID *string
	if u.BranchID != "" {
		branchID = &u.BranchID
	}
	_, err := r.db.Exec(ctx,
		`UPDATE users SET first_name=$2, last_name=$3, phone=$4, role=$5, branch_id=$6, updated_at=NOW()
		 WHERE id=$1`,
		u.ID, u.FirstName, u.LastName, u.Phone, u.Role, branchID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id, firstName, lastName, phone string, addr domain.Address) (*domain.User, error) {
	addrJSON, _ := json.Marshal(addr)
	return r.scan(r.db.QueryRow(ctx, `
		UPDATE users
		SET first_name=$2, last_name=$3, phone=$4, default_address=$5, updated_at=NOW()
		WHERE id=$1
		RETURNING `+userCols,
		id, firstName, lastName, phone, addrJSON))
}

func (r *UserRepo) ChangePassword(ctx context.Context, id, newHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1`, id, newHash)
	return err
}

func (r *UserRepo) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_active=$2, updated_at=NOW() WHERE id=$1`, id, active)
	return err
}

func (r *UserRepo) SetMFASecret(ctx context.Context, id, secret string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET mfa_secret=$2, mfa_enabled=true, updated_at=NOW() WHERE id=$1`, id, secret)
	return err
}

func (r *UserRepo) List(ctx context.Context, page, pageSize int) (*ports.Page[domain.User], error) {
	var total int64
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx,
		`SELECT `+userCols+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		u, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &ports.Page[domain.User]{
		Data: users, Total: total, Page: page, PageSize: pageSize,
		TotalPages: totalPages, HasNext: page < totalPages, HasPrev: page > 1,
	}, rows.Err()
}

func (r *UserRepo) scan(row rowScanner) (*domain.User, error) {
	var u domain.User
	var addrJSON []byte
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.BranchID,
		&u.FirstName, &u.LastName, &u.Phone, &addrJSON,
		&u.MFASecret, &u.MFAEnabled, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	json.Unmarshal(addrJSON, &u.DefaultAddress)
	return &u, nil
}
