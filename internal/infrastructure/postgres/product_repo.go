package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/ports"
)

type ProductRepo struct{ db *pgxpool.Pool }

func NewProductRepo(db *pgxpool.Pool) *ProductRepo { return &ProductRepo{db} }

var sortMap = map[string]string{
	"name":   "p.name ASC",
	"-name":  "p.name DESC",
	"price":  "COALESCE(bp.price, p.base_price) ASC",
	"-price": "COALESCE(bp.price, p.base_price) DESC",
}

func (r *ProductRepo) List(ctx context.Context, f domain.ProductFilter) (*ports.Page[domain.Product], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 20
	}

	order := "p.name ASC"
	if s, ok := sortMap[f.SortBy]; ok {
		order = s
	}

	args := []any{}
	where := []string{"p.is_active = true"}
	i := 1

	branchArg := func() string { return fmt.Sprintf("$%d", i) }

	// Branch ID for price join
	var branchJoin string
	if f.BranchID != "" {
		branchJoin = fmt.Sprintf("LEFT JOIN branch_prices bp ON bp.product_id = p.id AND bp.branch_id = $%d\n\t\tLEFT JOIN inventory inv ON inv.product_id = p.id AND inv.branch_id = $%d", i, i)
		args = append(args, f.BranchID)
		i++
	} else {
		branchJoin = "LEFT JOIN branch_prices bp ON false\n\t\tLEFT JOIN (SELECT product_id, SUM(quantity-reserved_qty) AS quantity, 0 AS reserved_qty FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id"
		_ = branchArg
	}

	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("p.category_id = $%d", i))
		args = append(args, f.CategoryID)
		i++
	}

	if f.Search != "" {
		where = append(where, fmt.Sprintf("p.search_vector @@ plainto_tsquery('spanish', $%d)", i))
		args = append(args, f.Search)
		i++
	}

	if f.InStock {
		where = append(where, "COALESCE(inv.quantity - inv.reserved_qty, 0) > 0")
	}

	if f.PriceMin > 0 {
		where = append(where, fmt.Sprintf("COALESCE(bp.price, p.base_price) >= $%d", i))
		args = append(args, f.PriceMin)
		i++
	}

	if f.PriceMax > 0 {
		where = append(where, fmt.Sprintf("COALESCE(bp.price, p.base_price) <= $%d", i))
		args = append(args, f.PriceMax)
		i++
	}

	if f.Tag != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(p.tags)", i))
		args = append(args, f.Tag)
		i++
	}

	whereClause := strings.Join(where, " AND ")

	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM products p %s WHERE %s`, branchJoin, whereClause)
	var total int64
	r.db.QueryRow(ctx, countSQL, args...).Scan(&total)

	offset := (f.Page - 1) * f.PageSize
	limitArg := fmt.Sprintf("$%d", i)
	offsetArg := fmt.Sprintf("$%d", i+1)
	args = append(args, f.PageSize, offset)

	dataSQL := fmt.Sprintf(`
		SELECT p.id, p.sku, p.name, p.description, p.base_price,
		       bp.price,
		       COALESCE(p.category_id::text,''),
		       p.images, p.attributes, p.tags, p.is_active,
		       COALESCE(inv.quantity - inv.reserved_qty, 0),
		       p.created_at, p.updated_at
		FROM products p
		%s
		WHERE %s
		ORDER BY %s
		LIMIT %s OFFSET %s`,
		branchJoin, whereClause, order, limitArg, offsetArg)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *p)
	}

	totalPages := int(math.Ceil(float64(total) / float64(f.PageSize)))
	return &ports.Page[domain.Product]{
		Data: products, Total: total, Page: f.Page, PageSize: f.PageSize,
		TotalPages: totalPages, HasNext: f.Page < totalPages, HasPrev: f.Page > 1,
	}, rows.Err()
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return r.scanProduct(r.db.QueryRow(ctx, `
		SELECT p.id, p.sku, p.name, p.description, p.base_price,
		       NULL::numeric,
		       COALESCE(p.category_id::text,''),
		       p.images, p.attributes, p.tags, p.is_active,
		       COALESCE((SELECT SUM(quantity - reserved_qty) FROM inventory WHERE product_id = p.id), 0),
		       p.created_at, p.updated_at
		FROM products p WHERE p.id = $1`, id))
}

func (r *ProductRepo) GetWithPrice(ctx context.Context, id, branchID string) (*domain.Product, error) {
	return r.scanProduct(r.db.QueryRow(ctx, `
		SELECT p.id, p.sku, p.name, p.description, p.base_price,
		       bp.price,
		       COALESCE(p.category_id::text,''),
		       p.images, p.attributes, p.tags, p.is_active,
		       COALESCE(inv.quantity - inv.reserved_qty, 0),
		       p.created_at, p.updated_at
		FROM products p
		LEFT JOIN branch_prices bp ON bp.product_id = p.id AND bp.branch_id = $2
		LEFT JOIN inventory inv ON inv.product_id = p.id AND inv.branch_id = $2
		WHERE p.id = $1`, id, branchID))
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	images, _ := json.Marshal(p.Images)
	attrs, _ := json.Marshal(p.Attributes)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	var catID *string
	if p.CategoryID != "" {
		catID = &p.CategoryID
	}
	return r.scanProduct(r.db.QueryRow(ctx, `
		INSERT INTO products (id, sku, name, description, base_price, category_id, images, attributes, tags, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, sku, name, description, base_price, NULL::numeric,
		          COALESCE(category_id::text,''), images, attributes, tags, is_active, 0, created_at, updated_at`,
		p.ID, p.SKU, p.Name, p.Description, p.BasePrice, catID,
		images, attrs, p.Tags, p.IsActive))
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	images, _ := json.Marshal(p.Images)
	attrs, _ := json.Marshal(p.Attributes)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	var catID *string
	if p.CategoryID != "" {
		catID = &p.CategoryID
	}
	_, err := r.db.Exec(ctx, `
		UPDATE products
		SET sku=$2, name=$3, description=$4, base_price=$5, category_id=$6,
		    images=$7, attributes=$8, tags=$9, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.SKU, p.Name, p.Description, p.BasePrice, catID,
		images, attrs, p.Tags)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.Exec(ctx, `UPDATE products SET is_active=$2, updated_at=NOW() WHERE id=$1`, id, active)
	return err
}

func (r *ProductRepo) SetBranchPrice(ctx context.Context, op domain.OverridePrice) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO branch_prices (product_id, branch_id, price)
		VALUES ($1,$2,$3)
		ON CONFLICT (product_id, branch_id) DO UPDATE SET price=$3, updated_at=NOW()`,
		op.ProductID, op.BranchID, op.Price)
	return err
}

func (r *ProductRepo) BulkUpsert(ctx context.Context, products []*domain.Product) (int, error) {
	count := 0
	for _, p := range products {
		if p.ID == "" {
			p.ID = uuid.New().String()
		}
		images, _ := json.Marshal(p.Images)
		attrs, _ := json.Marshal(p.Attributes)
		if p.Tags == nil {
			p.Tags = []string{}
		}
		var catID *string
		if p.CategoryID != "" {
			catID = &p.CategoryID
		}
		_, err := r.db.Exec(ctx, `
			INSERT INTO products (id, sku, name, description, base_price, category_id, images, attributes, tags, is_active)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (sku) DO UPDATE
			SET name=$3, description=$4, base_price=$5, category_id=$6,
			    images=$7, attributes=$8, tags=$9, updated_at=NOW()`,
			p.ID, p.SKU, p.Name, p.Description, p.BasePrice, catID,
			images, attrs, p.Tags, p.IsActive)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

const catCols = `c.id, c.name, c.slug, c.parent_id,
	COALESCE(c.description,''), COALESCE(c.sort_order,0), COALESCE(c.is_active,true),
	COUNT(p.id)`

func (r *ProductRepo) scanCategory(row rowScanner) (*domain.Category, error) {
	var c domain.Category
	if err := row.Scan(
		&c.ID, &c.Name, &c.Slug, &c.ParentID,
		&c.Description, &c.SortOrder, &c.IsActive, &c.ProductCount,
	); err != nil {
		return nil, err
	}
	c.Children = []domain.Category{}
	return &c, nil
}

func (r *ProductRepo) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+catCols+`
		FROM categories c
		LEFT JOIN products p ON p.category_id = c.id AND p.is_active = true
		GROUP BY c.id
		ORDER BY c.sort_order, c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	all := map[string]*domain.Category{}
	order := []string{}
	for rows.Next() {
		c, err := r.scanCategory(rows)
		if err != nil {
			return nil, err
		}
		all[c.ID] = c
		order = append(order, c.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]domain.Category, 0, len(order))
	for _, id := range order {
		result = append(result, *all[id])
	}
	return result, nil
}

func (r *ProductRepo) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return r.scanCategory(r.db.QueryRow(ctx, `
		SELECT `+catCols+`
		FROM categories c
		LEFT JOIN products p ON p.category_id = c.id AND p.is_active = true
		WHERE c.id=$1
		GROUP BY c.id`, id))
}

func (r *ProductRepo) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	var parentID *string
	if c.ParentID != nil && *c.ParentID != "" {
		parentID = c.ParentID
	}
	return r.scanCategory(r.db.QueryRow(ctx, `
		INSERT INTO categories (name, slug, parent_id, description, sort_order)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, name, slug, parent_id,
		          COALESCE(description,''), COALESCE(sort_order,0), COALESCE(is_active,true), 0`,
		c.Name, c.Slug, parentID, c.Description, c.SortOrder))
}

func (r *ProductRepo) UpdateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	var parentID *string
	if c.ParentID != nil && *c.ParentID != "" {
		parentID = c.ParentID
	}
	return r.scanCategory(r.db.QueryRow(ctx, `
		UPDATE categories
		SET name=$2, slug=$3, parent_id=$4, description=$5, sort_order=$6
		WHERE id=$1
		RETURNING id, name, slug, parent_id,
		          COALESCE(description,''), COALESCE(sort_order,0), COALESCE(is_active,true), 0`,
		c.ID, c.Name, c.Slug, parentID, c.Description, c.SortOrder))
}

func (r *ProductRepo) SetCategoryActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE categories SET is_active=$2 WHERE id=$1`, id, active)
	return err
}

func (r *ProductRepo) ListTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT unnest(tags) AS tag
		FROM products
		WHERE is_active = true AND array_length(tags, 1) > 0
		ORDER BY tag`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, rows.Err()
}

func (r *ProductRepo) scanProduct(row rowScanner) (*domain.Product, error) {
	var p domain.Product
	var images, attrs []byte
	var branchPrice *float64
	if err := row.Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.BasePrice,
		&branchPrice, &p.CategoryID,
		&images, &attrs, &p.Tags, &p.IsActive, &p.Stock,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	p.BranchPrice = branchPrice
	json.Unmarshal(images, &p.Images)
	json.Unmarshal(attrs, &p.Attributes)
	if p.Images == nil {
		p.Images = []domain.ProductImage{}
	}
	if p.Attributes == nil {
		p.Attributes = map[string]any{}
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	return &p, nil
}
