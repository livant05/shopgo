// Package middleware — limpio, sin TenantResolver, sin search_path dinámico.
package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims embebido localmente para no crear dependencias circulares.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string `json:"uid"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	BranchID string `json:"bid,omitempty"`
}

// Logger — structured JSON logging con slog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rid := uuid.New().String()
		c.Set("request_id", rid)
		c.Header("X-Request-ID", rid)

		c.Next()

		slog.Info("req",
			"method",     c.Request.Method,
			"path",       c.Request.URL.Path,
			"status",     c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip",         c.ClientIP(),
			"request_id", rid,
		)
	}
}

// Recovery — captura panics, retorna 500 limpio.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic", "err", r, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "error interno del servidor",
				})
			}
		}()
		c.Next()
	}
}

// SecurityHeaders — headers HTTP de seguridad mínimos pero correctos.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// JWT — verifica y parsea el token. Inyecta claims al contexto Gin.
func JWT(pubKey any) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "MISSING_TOKEN"})
			return
		}

		claims := &Claims{}
		_, err := jwt.ParseWithClaims(strings.TrimPrefix(auth, "Bearer "), claims, func(_ *jwt.Token) (any, error) {
			return pubKey, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "INVALID_TOKEN", "message": err.Error()})
			return
		}

		c.Set("user_id",   claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("branch_id", claims.BranchID)
		c.Next()
	}
}

// RequireRole — RBAC simple con jerarquía numérica.
func RequireRole(min string) gin.HandlerFunc {
	levels := map[string]int{
		"admin": 100, "manager": 60, "staff": 40, "customer": 10,
	}
	return func(c *gin.Context) {
		if levels[c.GetString("user_role")] < levels[min] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "permisos insuficientes",
			})
			return
		}
		c.Next()
	}
}

// BranchScope — el staff de una sucursal solo puede ver su sucursal.
// Los admin/manager pueden operar en cualquier sucursal.
func BranchScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		role     := c.GetString("user_role")
		branchID := c.GetString("branch_id")

		// Admins ven todo — no limitar
		if role == "admin" {
			c.Next()
			return
		}
		// Staff y managers solo ven su sucursal asignada
		if branchID != "" {
			c.Set("scoped_branch", branchID)
		}
		c.Next()
	}
}

// IPWhitelist — solo para rutas de administración interna.
func IPWhitelist(allowed []string) gin.HandlerFunc {
	if len(allowed) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	set := make(map[string]bool, len(allowed))
	for _, ip := range allowed {
		set[strings.TrimSpace(ip)] = true
	}
	return func(c *gin.Context) {
		if !set[c.ClientIP()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "IP_NOT_ALLOWED"})
			return
		}
		c.Next()
	}
}

// ErrorHandler — serializa errores no manejados.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"code": appErr.Code, "message": appErr.Message})
			return
		}
		slog.Error("error no manejado", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "error interno"})
	}
}

// AppError — error estructurado de la aplicación.
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
