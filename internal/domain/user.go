package domain

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"    // acceso total al sistema
	RoleManager  Role = "manager"  // su sucursal asignada
	RoleStaff    Role = "staff"    // POS y órdenes
	RoleCustomer Role = "customer" // storefront únicamente
)

var roleLevels = map[Role]int{
	RoleAdmin: 100, RoleManager: 60, RoleStaff: 40, RoleCustomer: 10,
}

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	Role           Role      `json:"role"`
	BranchID       string    `json:"branch_id,omitempty"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Phone          string    `json:"phone,omitempty"`
	DefaultAddress Address   `json:"default_address,omitempty"`
	MFASecret      string    `json:"-"`
	MFAEnabled     bool      `json:"mfa_enabled"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *User) HasRole(min Role) bool {
	return roleLevels[u.Role] >= roleLevels[min]
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
