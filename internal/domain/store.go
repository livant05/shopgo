// Package domain — entidades del negocio. Cero imports externos.
package domain

import "time"

// Branch es la unidad operativa. Una tienda, un almacén.
type Branch struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Address       Address        `json:"address"`
	WarehouseMode bool           `json:"warehouse_mode"`
	Settings      BranchSettings `json:"settings"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
}

type BranchSettings struct {
	TaxRate   float64 `json:"tax_rate"`
	Currency  string  `json:"currency"`
	OpenHours string  `json:"open_hours,omitempty"`
}

type Address struct {
	Street  string  `json:"street"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Zip     string  `json:"zip"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}

// StoreConfig — configuración global. Una sola fila en la BD.
type StoreConfig struct {
	StoreName       string  `json:"store_name"`
	LogoURL         string  `json:"logo_url,omitempty"`
	Currency        string  `json:"currency"`
	TaxRate         float64 `json:"tax_rate"`
	TaxInclusive    bool    `json:"tax_inclusive"`
	ContactEmail    string  `json:"contact_email"`
	SupportPhone    string  `json:"support_phone,omitempty"`
	StripePublicKey string  `json:"stripe_public_key,omitempty"`
	SocialInstagram string  `json:"social_instagram,omitempty"`
	SocialFacebook  string  `json:"social_facebook,omitempty"`
	SocialWhatsapp  string  `json:"social_whatsapp,omitempty"`
}
