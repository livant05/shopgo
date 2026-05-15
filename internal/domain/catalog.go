package domain

import "time"

type Product struct {
	ID          string         `json:"id"`
	SKU         string         `json:"sku"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	BasePrice   float64        `json:"base_price"`
	BranchPrice *float64       `json:"branch_price,omitempty"`
	CategoryID  string         `json:"category_id,omitempty"`
	Images      []ProductImage `json:"images"`
	Attributes  map[string]any `json:"attributes"`
	Tags        []string       `json:"tags"`
	IsActive    bool           `json:"is_active"`
	Stock       int            `json:"stock,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (p *Product) EffectivePrice() float64 {
	if p.BranchPrice != nil {
		return *p.BranchPrice
	}
	return p.BasePrice
}

type ProductImage struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
	Order   int    `json:"order"`
	IsMain  bool   `json:"is_main"`
}

type OverridePrice struct {
	ProductID string  `json:"product_id"`
	BranchID  string  `json:"branch_id"`
	Price     float64 `json:"price"`
}

type Category struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Slug     string     `json:"slug"`
	ParentID *string    `json:"parent_id,omitempty"`
	Children []Category `json:"children,omitempty"`
}

type ProductFilter struct {
	BranchID   string
	CategoryID string
	Search     string
	SortBy     string
	InStock    bool
	Page       int
	PageSize   int
}
