// Package handlers — thin HTTP layer. Bind → validate → call service → respond.
package handlers

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/shopgo/internal/domain"
	"github.com/yourorg/shopgo/internal/infrastructure/postgres"
	"github.com/yourorg/shopgo/internal/ports"
	"github.com/yourorg/shopgo/internal/services"
)

// ──── helpers ─────────────────────────────────────────────

func queryInt(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func queryFloat(c *gin.Context, key string, def float64) float64 {
	v, err := strconv.ParseFloat(c.Query(key), 64)
	if err != nil || v < 0 {
		return def
	}
	return v
}

func apiErr(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"code": code, "message": msg})
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		apiErr(c, http.StatusNotFound, "NOT_FOUND", "recurso no encontrado")
	case errors.Is(err, ports.ErrConflict):
		apiErr(c, http.StatusConflict, "CONFLICT", "ya existe un registro con ese identificador")
	case errors.Is(err, ports.ErrInsufficientStock):
		apiErr(c, http.StatusConflict, "INSUFFICIENT_STOCK", "stock insuficiente")
	case errors.Is(err, ports.ErrInvalidTransition):
		apiErr(c, http.StatusUnprocessableEntity, "INVALID_TRANSITION", "transición de estado inválida")
	case errors.Is(err, services.ErrInvalidCredentials):
		apiErr(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales incorrectas")
	case errors.Is(err, services.ErrAccountDisabled):
		apiErr(c, http.StatusForbidden, "ACCOUNT_DISABLED", "cuenta deshabilitada")
	case errors.Is(err, services.ErrMFARequired):
		apiErr(c, http.StatusUnauthorized, "MFA_REQUIRED", "se requiere código 2FA")
	case errors.Is(err, services.ErrInvalidMFA):
		apiErr(c, http.StatusUnauthorized, "INVALID_MFA", "código 2FA incorrecto")
	default:
		apiErr(c, http.StatusInternalServerError, "INTERNAL_ERROR", "error interno del servidor")
	}
}

// ──── HealthHandler ────────────────────────────────────────

type HealthHandler struct {
	dbPing    func() error
	redisPing func() error
}

func NewHealthHandler(db, redis func() error) *HealthHandler {
	return &HealthHandler{db, redis}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := gin.H{}
	ok := true
	if err := h.dbPing(); err != nil {
		checks["postgres"] = err.Error()
		ok = false
	} else {
		checks["postgres"] = "ok"
	}
	if err := h.redisPing(); err != nil {
		checks["redis"] = err.Error()
		ok = false
	} else {
		checks["redis"] = "ok"
	}
	status := http.StatusOK
	if !ok {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"ok": ok, "checks": checks})
}

// ──── AuthHandler ──────────────────────────────────────────

type AuthHandler struct {
	auth  *services.AuthService
	users ports.UserRepository
}

func NewAuthHandler(auth *services.AuthService, users ports.UserRepository) *AuthHandler {
	return &AuthHandler{auth: auth, users: users}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		TOTPCode string `json:"totp_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	pair, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, req.TOTPCode)
	if err != nil {
		mapErr(c, err)
		return
	}
	user, _ := h.users.GetByEmail(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    pair.ExpiresIn,
		"user":          user,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", "refresh_token requerido")
		return
	}
	pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		apiErr(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error())
		return
	}
	c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email     string `json:"email"      binding:"required,email"`
		Password  string `json:"password"   binding:"required,min=8"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	hash, err := h.auth.HashPassword(req.Password)
	if err != nil {
		apiErr(c, http.StatusInternalServerError, "HASH_ERROR", "error interno")
		return
	}
	u := &domain.User{
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         domain.RoleCustomer,
		IsActive:     true,
	}
	created, err := h.users.Create(c.Request.Context(), u)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *AuthHandler) SetupMFA(c *gin.Context) {
	userID := c.GetString("user_id")
	secret, qrURL, err := h.auth.SetupMFA(c.Request.Context(), userID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"secret": secret, "qr_url": qrURL})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		FirstName string         `json:"first_name" binding:"required"`
		LastName  string         `json:"last_name"`
		Phone     string         `json:"phone"`
		Address   domain.Address `json:"default_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	user, err := h.users.UpdateProfile(c.Request.Context(),
		c.GetString("user_id"), req.FirstName, req.LastName, req.Phone, req.Address)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		mapErr(c, err)
		return
	}
	if err := h.auth.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		apiErr(c, http.StatusUnauthorized, "WRONG_PASSWORD", "contraseña actual incorrecta")
		return
	}
	hash, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		apiErr(c, http.StatusInternalServerError, "HASH_ERROR", "error interno")
		return
	}
	if err := h.users.ChangePassword(c.Request.Context(), user.ID, hash); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ──── StoreHandler ──────────────────────────────────────────

type storeRepository interface {
	Get(ctx context.Context) (*domain.StoreConfig, error)
	Update(ctx context.Context, sc *domain.StoreConfig) error
}

type StoreHandler struct{ repo storeRepository }

func NewStoreHandler(repo storeRepository) *StoreHandler {
	return &StoreHandler{repo}
}

func (h *StoreHandler) GetConfig(c *gin.Context) {
	sc, err := h.repo.Get(c.Request.Context())
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sc)
}

func (h *StoreHandler) UpdateConfig(c *gin.Context) {
	var sc domain.StoreConfig
	if err := c.ShouldBindJSON(&sc); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.repo.Update(c.Request.Context(), &sc); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sc)
}

// ──── ProductHandler ───────────────────────────────────────

type ProductHandler struct{ repo ports.ProductRepository }

func NewProductHandler(repo ports.ProductRepository) *ProductHandler {
	return &ProductHandler{repo}
}

func (h *ProductHandler) List(c *gin.Context) {
	f := domain.ProductFilter{
		BranchID:   c.Query("branch_id"),
		CategoryID: c.Query("category_id"),
		Search:     c.Query("q"),
		SortBy:     c.DefaultQuery("sort", "name"),
		InStock:    c.Query("in_stock") == "true",
		PriceMin:   queryFloat(c, "price_min", 0),
		PriceMax:   queryFloat(c, "price_max", 0),
		Tag:        c.Query("tag"),
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 20),
	}
	page, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *ProductHandler) Get(c *gin.Context) {
	branchID := c.Query("branch_id")
	var p *domain.Product
	var err error
	if branchID != "" {
		p, err = h.repo.GetWithPrice(c.Request.Context(), c.Param("id"), branchID)
	} else {
		p, err = h.repo.GetByID(c.Request.Context(), c.Param("id"))
	}
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	created, err := h.repo.Create(c.Request.Context(), &p)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	p.ID = c.Param("id")
	updated, err := h.repo.Update(c.Request.Context(), &p)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ProductHandler) SetActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	c.ShouldBindJSON(&req)
	if err := h.repo.SetActive(c.Request.Context(), c.Param("id"), req.IsActive); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ProductHandler) SetBranchPrice(c *gin.Context) {
	var req struct {
		BranchID string  `json:"branch_id" binding:"required"`
		Price    float64 `json:"price"     binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.repo.SetBranchPrice(c.Request.Context(), domain.OverridePrice{
		ProductID: c.Param("id"), BranchID: req.BranchID, Price: req.Price,
	}); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ProductHandler) BulkImport(c *gin.Context) {
	var products []*domain.Product
	if err := c.ShouldBindJSON(&products); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	n, err := h.repo.BulkUpsert(c.Request.Context(), products)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"upserted": n})
}

func (h *ProductHandler) ListTags(c *gin.Context) {
	tags, err := h.repo.ListTags(c.Request.Context())
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	cats, err := h.repo.ListCategories(c.Request.Context())
	if err != nil {
		mapErr(c, err)
		return
	}
	if cats == nil {
		cats = []domain.Category{}
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var body struct {
		Name        string  `json:"name" binding:"required"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		apiErr(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	slug := body.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(body.Name, " ", "-"))
	}
	cat := &domain.Category{
		Name:        body.Name,
		Slug:        slug,
		Description: body.Description,
		ParentID:    body.ParentID,
		SortOrder:   body.SortOrder,
		IsActive:    true,
	}
	created, err := h.repo.CreateCategory(c.Request.Context(), cat)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Name        string  `json:"name" binding:"required"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		apiErr(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	slug := body.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(body.Name, " ", "-"))
	}
	cat := &domain.Category{
		ID:          id,
		Name:        body.Name,
		Slug:        slug,
		Description: body.Description,
		ParentID:    body.ParentID,
		SortOrder:   body.SortOrder,
	}
	updated, err := h.repo.UpdateCategory(c.Request.Context(), cat)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ProductHandler) SetCategoryActive(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		apiErr(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if err := h.repo.SetCategoryActive(c.Request.Context(), id, body.Active); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ──── InventoryHandler ─────────────────────────────────────

type InventoryHandler struct{ repo ports.InventoryRepository }

func NewInventoryHandler(repo ports.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo}
}

func (h *InventoryHandler) List(c *gin.Context) {
	branchID := c.GetString("scoped_branch")
	if branchID == "" {
		branchID = c.Query("branch_id")
	}
	list, err := h.repo.List(c.Request.Context(), branchID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *InventoryHandler) ListAll(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context(), "")
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *InventoryHandler) LowStockAlerts(c *gin.Context) {
	branchID := c.GetString("scoped_branch")
	list, err := h.repo.GetLowStock(c.Request.Context(), branchID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "count": len(list)})
}

func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		BranchID  string `json:"branch_id"  binding:"required"`
		Delta     int    `json:"delta"      binding:"required"`
		Reason    string `json:"reason"     binding:"required"`
		Note      string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	userID := c.GetString("user_id")
	inv, err := h.repo.Adjust(c.Request.Context(), req.ProductID, req.BranchID, req.Delta, req.Reason, req.Note, userID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *InventoryHandler) Transfer(c *gin.Context) {
	var req struct {
		ProductID    string `json:"product_id"     binding:"required"`
		FromBranchID string `json:"from_branch_id" binding:"required"`
		ToBranchID   string `json:"to_branch_id"   binding:"required"`
		Quantity     int    `json:"quantity"       binding:"required,min=1"`
		Note         string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	userID := c.GetString("user_id")
	err := h.repo.Transfer(c.Request.Context(), ports.TransferCmd{
		ProductID:    req.ProductID,
		FromBranchID: req.FromBranchID,
		ToBranchID:   req.ToBranchID,
		Quantity:     req.Quantity,
		Note:         req.Note,
		UserID:       userID,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *InventoryHandler) History(c *gin.Context) {
	branchID := c.Query("branch_id")
	if branchID == "" {
		branchID = c.GetString("scoped_branch")
	}
	data, total, err := h.repo.History(
		c.Request.Context(),
		branchID,
		c.Query("type"),
		c.Query("from"),
		c.Query("to"),
		queryInt(c, "page", 1),
		queryInt(c, "page_size", 50),
	)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "total": total})
}

// ──── OrderHandler ─────────────────────────────────────────

type OrderHandler struct {
	svc  *services.OrderService
	repo ports.OrderRepository
}

func NewOrderHandler(svc *services.OrderService, repo ports.OrderRepository) *OrderHandler {
	return &OrderHandler{svc: svc, repo: repo}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req struct {
		BranchID string `json:"branch_id" binding:"required"`
		Items    []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int    `json:"quantity"   binding:"required,min=1"`
		} `json:"items" binding:"required,min=1"`
		CouponCode      string         `json:"coupon_code"`
		ShippingAddress domain.Address `json:"shipping_address"`
		Currency        string         `json:"currency"`
		Notes           string         `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	items := make([]services.OrderItemInput, len(req.Items))
	for i, it := range req.Items {
		items[i] = services.OrderItemInput{ProductID: it.ProductID, Quantity: it.Quantity}
	}

	order, err := h.svc.Create(c.Request.Context(), services.CreateOrderInput{
		BranchID:        req.BranchID,
		CustomerID:      c.GetString("user_id"),
		Items:           items,
		CouponCode:      req.CouponCode,
		ShippingAddress: req.ShippingAddress,
		Currency:        req.Currency,
		Notes:           req.Notes,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) Get(c *gin.Context) {
	order, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		mapErr(c, err)
		return
	}
	// Customers can only see their own orders
	if c.GetString("user_role") == "customer" && order.CustomerID != c.GetString("user_id") {
		apiErr(c, http.StatusForbidden, "FORBIDDEN", "acceso denegado")
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	f := ports.OrderFilter{
		CustomerID: c.GetString("user_id"),
		Status:     c.Query("status"),
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 20),
	}
	page, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *OrderHandler) ListBranch(c *gin.Context) {
	f := ports.OrderFilter{
		BranchID: c.GetString("scoped_branch"),
		Status:   c.Query("status"),
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 50),
	}
	page, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), req.Status); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *OrderHandler) ListAll(c *gin.Context) {
	f := ports.OrderFilter{
		Status:       c.Query("status"),
		RefundStatus: c.Query("refund_status"),
		Page:         queryInt(c, "page", 1),
		PageSize:     queryInt(c, "page_size", 20),
	}
	page, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *OrderHandler) RequestRefund(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.svc.RequestRefund(c.Request.Context(), c.Param("id"), req.Reason, c.GetString("user_id")); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *OrderHandler) ProcessRefund(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	var err error
	adminID := c.GetString("user_id")
	switch req.Action {
	case "approve":
		err = h.svc.ApproveRefund(c.Request.Context(), c.Param("id"), adminID)
	case "reject":
		err = h.svc.RejectRefund(c.Request.Context(), c.Param("id"))
	default:
		apiErr(c, http.StatusBadRequest, "INVALID_ACTION", "acción debe ser 'approve' o 'reject'")
		return
	}
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ──── PaymentHandler ───────────────────────────────────────

type PaymentHandler struct {
	svc     ports.PaymentService
	coupons couponValidatorFn
}

type couponValidatorFn func(ctx context.Context, code string, subtotal float64) (*domain.Coupon, float64, error)

func NewPaymentHandler(svc ports.PaymentService, validateCoupon couponValidatorFn) *PaymentHandler {
	return &PaymentHandler{svc: svc, coupons: validateCoupon}
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	var req struct {
		OrderID  string  `json:"order_id"  binding:"required"`
		Amount   float64 `json:"amount"    binding:"required,min=0.01"`
		Currency string  `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.Currency == "" {
		req.Currency = "MXN"
	}
	secret, err := h.svc.CreateIntent(c.Request.Context(), req.OrderID, req.Amount, strings.ToLower(req.Currency))
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_secret": secret})
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	payload, _ := c.GetRawData()
	sig := c.GetHeader("Stripe-Signature")
	if err := h.svc.HandleWebhook(c.Request.Context(), payload, sig); err != nil {
		apiErr(c, http.StatusBadRequest, "WEBHOOK_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *PaymentHandler) ValidateCoupon(c *gin.Context) {
	var req struct {
		Code     string  `json:"code"     binding:"required"`
		Subtotal float64 `json:"subtotal" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}

	if h.coupons == nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "discount": 0})
		return
	}

	coupon, discount, err := h.coupons(c.Request.Context(), req.Code, req.Subtotal)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "discount": 0, "message": "cupón inválido o expirado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"discount": discount,
		"coupon":   coupon,
	})
}

func (h *PaymentHandler) GetConfig(c *gin.Context) {
	configured := h.svc != nil
	c.JSON(http.StatusOK, gin.H{"stripe_configured": configured})
}

func (h *PaymentHandler) UpdateConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ──── BranchHandler ────────────────────────────────────────

type BranchHandler struct {
	branches ports.BranchRepository
	users    ports.UserRepository
	auth     *services.AuthService
}

func NewBranchHandler(branches ports.BranchRepository, users ports.UserRepository, auth *services.AuthService) *BranchHandler {
	return &BranchHandler{branches: branches, users: users, auth: auth}
}

func (h *BranchHandler) List(c *gin.Context) {
	list, err := h.branches.List(c.Request.Context())
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BranchHandler) Create(c *gin.Context) {
	var b domain.Branch
	if err := c.ShouldBindJSON(&b); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	created, err := h.branches.Create(c.Request.Context(), &b)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *BranchHandler) Update(c *gin.Context) {
	var b domain.Branch
	if err := c.ShouldBindJSON(&b); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	b.ID = c.Param("id")
	updated, err := h.branches.Update(c.Request.Context(), &b)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *BranchHandler) SetActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	c.ShouldBindJSON(&req)
	if err := h.branches.SetActive(c.Request.Context(), c.Param("id"), req.IsActive); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *BranchHandler) ListUsers(c *gin.Context) {
	page, err := h.users.List(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 50))
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *BranchHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email     string      `json:"email"      binding:"required,email"`
		Password  string      `json:"password"   binding:"required,min=8"`
		Role      domain.Role `json:"role"       binding:"required"`
		BranchID  string      `json:"branch_id"`
		FirstName string      `json:"first_name"`
		LastName  string      `json:"last_name"`
		Phone     string      `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	hash, err := h.auth.HashPassword(req.Password)
	if err != nil {
		mapErr(c, err)
		return
	}
	user, err := h.users.Create(c.Request.Context(), &domain.User{
		Email: req.Email, PasswordHash: hash, Role: req.Role,
		BranchID: req.BranchID, FirstName: req.FirstName,
		LastName: req.LastName, Phone: req.Phone,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *BranchHandler) UpdateUser(c *gin.Context) {
	var req struct {
		FirstName string      `json:"first_name"`
		LastName  string      `json:"last_name"`
		Phone     string      `json:"phone"`
		Role      domain.Role `json:"role"`
		BranchID  string      `json:"branch_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	user, err := h.users.Update(c.Request.Context(), &domain.User{
		ID: c.Param("id"), FirstName: req.FirstName, LastName: req.LastName,
		Phone: req.Phone, Role: req.Role, BranchID: req.BranchID,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *BranchHandler) SetUserActive(c *gin.Context) {
	var req struct {
		IsActive bool `json:"is_active"`
	}
	c.ShouldBindJSON(&req)
	if err := h.users.SetActive(c.Request.Context(), c.Param("id"), req.IsActive); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ──── ReportHandler ────────────────────────────────────────

type ReportHandler struct{ repo ports.ReportRepository }

func NewReportHandler(repo ports.ReportRepository) *ReportHandler {
	return &ReportHandler{repo}
}

func (h *ReportHandler) Dashboard(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	m, err := h.repo.Revenue(c.Request.Context(), from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *ReportHandler) Sales(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.DailySeries(c.Request.Context(), c.Query("branch_id"), from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) TopProducts(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	n := queryInt(c, "n", 10)
	data, err := h.repo.TopProducts(c.Request.Context(), c.Query("branch_id"), from, to, n)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) TopCustomers(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	n := queryInt(c, "n", 10)
	data, err := h.repo.TopCustomers(c.Request.Context(), from, to, n)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) HourlySeries(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.HourlySeries(c.Request.Context(), c.Query("branch_id"), from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) ByBranch(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.SalesByBranch(c.Request.Context(), from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) Daily(c *gin.Context) {
	branchID := c.GetString("scoped_branch")
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.DailySeries(c.Request.Context(), branchID, from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) CashClose(c *gin.Context) {
	branchID := c.GetString("scoped_branch")
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.DailySeries(c.Request.Context(), branchID, from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ReportHandler) ExportSales(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	data, err := h.repo.DailySeries(c.Request.Context(), c.Query("branch_id"), from, to)
	if err != nil {
		mapErr(c, err)
		return
	}
	filename := fmt.Sprintf("ventas_%s_%s.csv", from, to)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"Fecha", "Órdenes", "Ingresos (MXN)"})
	for _, s := range data {
		_ = w.Write([]string{
			s.Day,
			strconv.FormatInt(s.Orders, 10),
			fmt.Sprintf("%.2f", s.Revenue),
		})
	}
	w.Flush()
}

func (h *ReportHandler) ExportInventory(c *gin.Context) {
	branchID := c.Query("branch_id")
	data, err := h.repo.InventoryReport(c.Request.Context(), branchID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=inventario.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"SKU", "Producto", "Sucursal", "Cantidad", "Reservado", "Disponible", "Punto reorden", "Stock bajo", "Precio"})
	for _, r := range data {
		low := "No"
		if r.IsLow {
			low = "Sí"
		}
		_ = w.Write([]string{
			r.SKU, r.ProductName, r.BranchName,
			strconv.Itoa(r.Quantity),
			strconv.Itoa(r.ReservedQty),
			strconv.Itoa(r.Available),
			strconv.Itoa(r.ReorderPoint),
			low,
			fmt.Sprintf("%.2f", r.Price),
		})
	}
	w.Flush()
}

// ──── CouponHandler ───────────────────────────────────────

type couponRepo interface {
	List(ctx context.Context) ([]domain.Coupon, error)
	Create(ctx context.Context, in postgres.CreateCouponInput) (*domain.Coupon, error)
	SetActive(ctx context.Context, id string, active bool) error
}

type CouponHandler struct{ repo couponRepo }

func NewCouponHandler(repo couponRepo) *CouponHandler { return &CouponHandler{repo} }

func (h *CouponHandler) List(c *gin.Context) {
	coupons, err := h.repo.List(c.Request.Context())
	if err != nil {
		mapErr(c, err)
		return
	}
	if coupons == nil {
		coupons = []domain.Coupon{}
	}
	c.JSON(http.StatusOK, gin.H{"data": coupons})
}

func (h *CouponHandler) Create(c *gin.Context) {
	var req postgres.CreateCouponInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.Code == "" || req.Type == "" || req.Value <= 0 {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", "code, type y value son requeridos")
		return
	}
	if req.Type != "percent" && req.Type != "fixed" {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", "type debe ser 'percent' o 'fixed'")
		return
	}
	coupon, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, coupon)
}

func (h *CouponHandler) SetActive(c *gin.Context) {
	var req struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if err := h.repo.SetActive(c.Request.Context(), c.Param("id"), req.Active); err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
