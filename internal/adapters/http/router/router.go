// Package router — rutas organizadas por rol, sin complejidad de tenant.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/shopgo/internal/adapters/http/handlers"
	"github.com/yourorg/shopgo/internal/adapters/http/middleware"
)

type Deps struct {
	Auth      *handlers.AuthHandler
	Product   *handlers.ProductHandler
	Inventory *handlers.InventoryHandler
	Order     *handlers.OrderHandler
	Report    *handlers.ReportHandler
	Payment   *handlers.PaymentHandler
	Branch    *handlers.BranchHandler
	Store     *handlers.StoreHandler
	Coupon    *handlers.CouponHandler
	Notify    *handlers.NotifyHandler
	Quote     *handlers.QuoteHandler
	Health    *handlers.HealthHandler
	PubKey    any
	AdminIPs  []string
}

func Setup(r *gin.Engine, d *Deps) {
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.SecurityHeaders())

	// ── Infraestructura ─────────────────────────────
	r.GET("/health", d.Health.Live)
	r.GET("/ready", d.Health.Ready)

	api := r.Group("/api/v1")

	// ── Públicas (sin auth) ──────────────────────────
	pub := api.Group("")
	{
		pub.POST("/auth/login", d.Auth.Login)
		pub.POST("/auth/register", d.Auth.Register)
		pub.POST("/auth/refresh", d.Auth.Refresh)

		pub.GET("/products", d.Product.List)
		pub.GET("/products/:id", d.Product.Get)
		pub.GET("/tags", d.Product.ListTags)
		pub.GET("/categories", d.Product.ListCategories)
		pub.POST("/quotes", d.Quote.Create)
		pub.GET("/quotes/:id", d.Quote.Get)
		pub.GET("/branches", d.Branch.List)
		pub.GET("/store", d.Store.GetConfig)

		pub.POST("/webhooks/stripe", d.Payment.Webhook)
	}

	// ── Autenticadas — customer+ ─────────────────────
	mw := middleware.JWT(d.PubKey)
	auth := api.Group("", mw)
	{
		auth.GET("/auth/me", d.Auth.Me)
		auth.POST("/orders", middleware.RequireRole("customer"), d.Order.Create)
		auth.GET("/orders/:id", middleware.RequireRole("customer"), d.Order.Get)
		auth.GET("/orders", middleware.RequireRole("customer"), d.Order.List)
		auth.POST("/orders/:id/refund", middleware.RequireRole("customer"), d.Order.RequestRefund)
		auth.PUT("/auth/profile", d.Auth.UpdateProfile)
		auth.PUT("/auth/password", d.Auth.ChangePassword)
		auth.POST("/payments/intent", middleware.RequireRole("customer"), d.Payment.CreateIntent)
		auth.POST("/coupons/validate", middleware.RequireRole("customer"), d.Payment.ValidateCoupon)
		auth.POST("/auth/setup-mfa", d.Auth.SetupMFA)
	}

	// ── Staff — operaciones de sucursal ──────────────
	staff := api.Group("/ops", mw, middleware.RequireRole("staff"), middleware.BranchScope())
	{
		staff.GET("/orders", d.Order.ListBranch)
		staff.PATCH("/orders/:id/status", d.Order.UpdateStatus)
		staff.GET("/inventory", d.Inventory.List)
		staff.GET("/inventory/alerts", d.Inventory.LowStockAlerts)
		staff.GET("/reports/daily", d.Report.Daily)
		staff.GET("/reports/cashclose", d.Report.CashClose)
	}

	// ── Manager — su sucursal + inventario ───────────
	mgr := api.Group("/mgr", mw, middleware.RequireRole("manager"), middleware.BranchScope())
	{
		mgr.PATCH("/inventory", d.Inventory.Adjust)
		mgr.POST("/inventory/transfer", d.Inventory.Transfer)
		mgr.GET("/inventory/history", d.Inventory.History)
	}

	// ── Admin — acceso total ──────────────────────────
	adm := api.Group("/admin", mw, middleware.RequireRole("admin"), middleware.IPWhitelist(d.AdminIPs))
	{
		// Productos
		adm.POST("/products", d.Product.Create)
		adm.PUT("/products/:id", d.Product.Update)
		adm.PATCH("/products/:id/active", d.Product.SetActive)
		adm.PUT("/products/:id/price", d.Product.SetBranchPrice)
		adm.POST("/products/bulk", d.Product.BulkImport)
		adm.POST("/categories", d.Product.CreateCategory)
		adm.PUT("/categories/:id", d.Product.UpdateCategory)
		adm.PATCH("/categories/:id/active", d.Product.SetCategoryActive)

		// Inventario total
		adm.GET("/inventory", d.Inventory.ListAll)
		adm.PATCH("/inventory", d.Inventory.Adjust)
		adm.POST("/inventory/transfer", d.Inventory.Transfer)
		adm.GET("/inventory/history", d.Inventory.History)

		// Sucursales
		adm.GET("/branches", d.Branch.List)
		adm.POST("/branches", d.Branch.Create)
		adm.PUT("/branches/:id", d.Branch.Update)
		adm.PATCH("/branches/:id/active", d.Branch.SetActive)

		// Usuarios
		adm.GET("/users", d.Branch.ListUsers)
		adm.POST("/users", d.Branch.CreateUser)
		adm.PUT("/users/:id", d.Branch.UpdateUser)
		adm.PATCH("/users/:id/active", d.Branch.SetUserActive)

		// Reportes
		adm.GET("/reports/dashboard", d.Report.Dashboard)
		adm.GET("/reports/sales", d.Report.Sales)
		adm.GET("/reports/products", d.Report.TopProducts)
		adm.GET("/reports/customers", d.Report.TopCustomers)
		adm.GET("/reports/hourly", d.Report.HourlySeries)
		adm.GET("/reports/branches", d.Report.ByBranch)
		adm.GET("/reports/export/sales", d.Report.ExportSales)
		adm.GET("/reports/export/inventory", d.Report.ExportInventory)

		// Notificaciones SSE (solo admin)
		adm.GET("/notifications/stream", d.Notify.Stream)

		// Cupones
		adm.GET("/coupons", d.Coupon.List)
		adm.POST("/coupons", d.Coupon.Create)
		adm.PATCH("/coupons/:id/active", d.Coupon.SetActive)

		// Órdenes admin (todas las sucursales)
		adm.GET("/orders", d.Order.ListAll)
		adm.PATCH("/orders/:id/status", d.Order.UpdateStatus)
		adm.PUT("/orders/:id/refund", d.Order.ProcessRefund)

		// Cotizaciones
		adm.GET("/quotes", d.Quote.List)

		// Configuración de la tienda
		adm.GET("/store", d.Store.GetConfig)
		adm.PUT("/store", d.Store.UpdateConfig)
		adm.GET("/store/payment", d.Payment.GetConfig)
		adm.PUT("/store/payment", d.Payment.UpdateConfig)
	}
}
