package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type reservationData struct {
	BranchID string              `json:"branch_id"`
	Items    []ports.ReserveItem `json:"items"`
}

type InventoryRepo struct {
	db    *pgxpool.Pool
	cache ports.CacheService
}

func NewInventoryRepo(db *pgxpool.Pool, cache ports.CacheService) *InventoryRepo {
	return &InventoryRepo{db: db, cache: cache}
}

func (r *InventoryRepo) Get(ctx context.Context, productID, branchID string) (*domain.Inventory, error) {
	var inv domain.Inventory
	err := r.db.QueryRow(ctx,
		`SELECT product_id, branch_id, quantity, reserved_qty, reorder_point, updated_at
		 FROM inventory WHERE product_id=$1 AND branch_id=$2`, productID, branchID).
		Scan(&inv.ProductID, &inv.BranchID, &inv.Quantity, &inv.ReservedQty, &inv.ReorderPoint, &inv.UpdatedAt)
	if err != nil {
		return nil, ports.ErrNotFound
	}
	return &inv, nil
}

func (r *InventoryRepo) List(ctx context.Context, branchID string) ([]*domain.Inventory, error) {
	var q string
	var args []any
	if branchID != "" {
		q = `SELECT product_id, branch_id, quantity, reserved_qty, reorder_point, updated_at
		     FROM inventory WHERE branch_id=$1 ORDER BY product_id`
		args = []any{branchID}
	} else {
		q = `SELECT product_id, branch_id, quantity, reserved_qty, reorder_point, updated_at
		     FROM inventory ORDER BY branch_id, product_id`
	}
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*domain.Inventory, 0)
	for rows.Next() {
		var inv domain.Inventory
		if err := rows.Scan(&inv.ProductID, &inv.BranchID, &inv.Quantity, &inv.ReservedQty, &inv.ReorderPoint, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *InventoryRepo) GetLowStock(ctx context.Context, branchID string) ([]*domain.Inventory, error) {
	q := `SELECT product_id, branch_id, quantity, reserved_qty, reorder_point, updated_at
	      FROM inventory WHERE (quantity - reserved_qty) <= reorder_point`
	args := []any{}
	if branchID != "" {
		q += " AND branch_id=$1"
		args = append(args, branchID)
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*domain.Inventory, 0)
	for rows.Next() {
		var inv domain.Inventory
		if err := rows.Scan(&inv.ProductID, &inv.BranchID, &inv.Quantity, &inv.ReservedQty, &inv.ReorderPoint, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *InventoryRepo) Adjust(ctx context.Context, productID, branchID string, delta int, reason, note, userID string) (*domain.Inventory, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var inv domain.Inventory
	err = tx.QueryRow(ctx, `
		INSERT INTO inventory (product_id, branch_id, quantity, reserved_qty, reorder_point)
		VALUES ($1,$2,GREATEST(0,$3),0,5)
		ON CONFLICT (product_id, branch_id) DO UPDATE
		SET quantity = GREATEST(0, inventory.quantity + $3), updated_at=NOW()
		RETURNING product_id, branch_id, quantity, reserved_qty, reorder_point, updated_at`,
		productID, branchID, delta).
		Scan(&inv.ProductID, &inv.BranchID, &inv.Quantity, &inv.ReservedQty, &inv.ReorderPoint, &inv.UpdatedAt)
	if err != nil {
		return nil, err
	}

	movType := "adjustment"
	if delta < 0 {
		movType = "reduction"
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, to_branch_id, quantity, type, reason, note, user_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		uuid.New().String(), productID, branchID, delta, movType, reason, note, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepo) Reserve(ctx context.Context, branchID string, items []ports.ReserveItem) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		res, err := tx.Exec(ctx, `
			UPDATE inventory
			SET reserved_qty = reserved_qty + $3, updated_at=NOW()
			WHERE product_id=$1 AND branch_id=$2
			  AND (quantity - reserved_qty) >= $3`,
			item.ProductID, branchID, item.Quantity)
		if err != nil {
			return "", err
		}
		if res.RowsAffected() == 0 {
			return "", ports.ErrInsufficientStock
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	reservationID := uuid.New().String()
	data := reservationData{BranchID: branchID, Items: items}
	r.cache.Set(ctx, fmt.Sprintf("reservation:%s", reservationID), data, int((30 * time.Minute).Seconds()))
	return reservationID, nil
}

func (r *InventoryRepo) Commit(ctx context.Context, reservationID string) error {
	var data reservationData
	if err := r.cache.Get(ctx, fmt.Sprintf("reservation:%s", reservationID), &data); err != nil {
		return nil // reservation already committed or expired
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range data.Items {
		_, err = tx.Exec(ctx, `
			UPDATE inventory
			SET quantity = quantity - $3, reserved_qty = reserved_qty - $3, updated_at=NOW()
			WHERE product_id=$1 AND branch_id=$2`,
			item.ProductID, data.BranchID, item.Quantity)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	r.cache.Delete(ctx, fmt.Sprintf("reservation:%s", reservationID))
	return nil
}

func (r *InventoryRepo) Release(ctx context.Context, reservationID string) error {
	var data reservationData
	key := fmt.Sprintf("reservation:%s", reservationID)
	if err := r.cache.Get(ctx, key, &data); err != nil {
		return nil // already released or expired
	}

	for _, item := range data.Items {
		r.db.Exec(ctx, `
			UPDATE inventory
			SET reserved_qty = GREATEST(0, reserved_qty - $3), updated_at=NOW()
			WHERE product_id=$1 AND branch_id=$2`,
			item.ProductID, data.BranchID, item.Quantity)
	}
	r.cache.Delete(ctx, key)
	return nil
}

func (r *InventoryRepo) Transfer(ctx context.Context, cmd ports.TransferCmd) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE inventory
		SET quantity = quantity - $3, updated_at=NOW()
		WHERE product_id=$1 AND branch_id=$2 AND (quantity - reserved_qty) >= $3`,
		cmd.ProductID, cmd.FromBranchID, cmd.Quantity)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ports.ErrInsufficientStock
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO inventory (product_id, branch_id, quantity, reserved_qty, reorder_point)
		VALUES ($1,$2,$3,0,5)
		ON CONFLICT (product_id, branch_id) DO UPDATE
		SET quantity = inventory.quantity + $3, updated_at=NOW()`,
		cmd.ProductID, cmd.ToBranchID, cmd.Quantity)
	if err != nil {
		return err
	}

	movType := "transfer"
	b, _ := json.Marshal(map[string]string{"from": cmd.FromBranchID, "to": cmd.ToBranchID})
	_ = b
	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_movements (id, product_id, from_branch_id, to_branch_id, quantity, type, reason, note, user_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		uuid.New().String(), cmd.ProductID, cmd.FromBranchID, cmd.ToBranchID,
		cmd.Quantity, movType, "transfer", cmd.Note, cmd.UserID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *InventoryRepo) History(ctx context.Context, branchID, movType, from, to string, page, pageSize int) ([]*domain.InventoryMovement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	where := []string{"1=1"}
	args := []any{}
	i := 1

	if branchID != "" {
		where = append(where, fmt.Sprintf("(m.from_branch_id=$%d OR m.to_branch_id=$%d)", i, i+1))
		args = append(args, branchID, branchID)
		i += 2
	}
	if movType != "" {
		where = append(where, fmt.Sprintf("m.type=$%d", i))
		args = append(args, movType)
		i++
	}
	if from != "" {
		where = append(where, fmt.Sprintf("m.created_at >= $%d::timestamptz", i))
		args = append(args, from)
		i++
	}
	if to != "" {
		where = append(where, fmt.Sprintf("m.created_at <= $%d::timestamptz", i))
		args = append(args, to)
		i++
	}

	cond := ""
	for j, w := range where {
		if j == 0 {
			cond = w
		} else {
			cond += " AND " + w
		}
	}

	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_movements m WHERE `+cond, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, `
		SELECT m.id, m.product_id, COALESCE(p.name,''), COALESCE(p.sku,''),
		       COALESCE(m.from_branch_id::text,''), COALESCE(fb.name,''),
		       COALESCE(m.to_branch_id::text,''),   COALESCE(tb.name,''),
		       m.quantity, m.type, m.reason, COALESCE(m.note,''),
		       m.user_id::text, m.created_at
		FROM inventory_movements m
		LEFT JOIN products  p  ON p.id  = m.product_id
		LEFT JOIN branches  fb ON fb.id = m.from_branch_id
		LEFT JOIN branches  tb ON tb.id = m.to_branch_id
		WHERE `+cond+`
		ORDER BY m.created_at DESC
		LIMIT $`+fmt.Sprintf("%d", i)+` OFFSET $`+fmt.Sprintf("%d", i+1),
		args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*domain.InventoryMovement, 0)
	for rows.Next() {
		var m domain.InventoryMovement
		if err := rows.Scan(
			&m.ID, &m.ProductID, &m.ProductName, &m.ProductSKU,
			&m.FromBranchID, &m.FromBranchName,
			&m.ToBranchID, &m.ToBranchName,
			&m.Quantity, &m.Type, &m.Reason, &m.Note,
			&m.UserID, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, &m)
	}
	return list, total, rows.Err()
}
