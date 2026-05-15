package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourorg/shopgo/internal/adapters/http/handlers"
	"github.com/yourorg/shopgo/internal/adapters/http/router"
	"github.com/yourorg/shopgo/internal/config"
	"github.com/yourorg/shopgo/internal/infrastructure/postgres"
	redisclient "github.com/yourorg/shopgo/internal/infrastructure/redis"
	stripesvc "github.com/yourorg/shopgo/internal/infrastructure/stripe"
	"github.com/yourorg/shopgo/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuración inválida", "err", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel)
	slog.Info("arrancando ShopGo", "env", cfg.AppEnv, "port", cfg.Port, "store", cfg.StoreName)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ── Infraestructura ──────────────────────────────────
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
	if err != nil {
		slog.Error("conectar PostgreSQL", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("PostgreSQL conectado", "max_conns", cfg.DBMaxConns)

	redisClient, err := redisclient.New(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		slog.Error("conectar Redis", "err", err)
		os.Exit(1)
	}
	slog.Info("Redis conectado")

	// ── JWT keys ─────────────────────────────────────────
	privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath)
	if err != nil {
		slog.Error("leer clave privada JWT", "path", cfg.JWTPrivateKeyPath, "err", err)
		os.Exit(1)
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		slog.Error("parsear clave privada JWT", "err", err)
		os.Exit(1)
	}
	pubPEM, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		slog.Error("leer clave pública JWT", "path", cfg.JWTPublicKeyPath, "err", err)
		os.Exit(1)
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		slog.Error("parsear clave pública JWT", "err", err)
		os.Exit(1)
	}

	// ── Repositorios ──────────────────────────────────────
	branchRepo    := postgres.NewBranchRepo(pool)
	userRepo      := postgres.NewUserRepo(pool)
	productRepo   := postgres.NewProductRepo(pool)
	inventoryRepo := postgres.NewInventoryRepo(pool, redisClient)
	orderRepo     := postgres.NewOrderRepo(pool)
	storeRepo     := postgres.NewStoreRepo(pool)
	couponRepo    := postgres.NewCouponRepo(pool)
	reportRepo    := postgres.NewReportRepo(pool)

	// ── Servicios ────────────────────────────────────────
	authSvc  := services.NewAuthService(userRepo, redisClient, privKey, pubKey, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	orderSvc := services.NewOrderService(orderRepo, inventoryRepo, productRepo, branchRepo, redisClient)
	paySvc   := stripesvc.NewPaymentService(cfg.StripeSecretKey, cfg.StripeWebhookSecret, orderRepo)

	// ── Handlers y router ────────────────────────────────
	r := gin.New()

	deps := &router.Deps{
		Health: handlers.NewHealthHandler(
			func() error { return pool.Ping(context.Background()) },
			func() error { return redisClient.Ping(context.Background()) },
		),
		Auth:      handlers.NewAuthHandler(authSvc, userRepo),
		Product:   handlers.NewProductHandler(productRepo),
		Inventory: handlers.NewInventoryHandler(inventoryRepo),
		Order:     handlers.NewOrderHandler(orderSvc, orderRepo),
		Report:    handlers.NewReportHandler(reportRepo),
		Payment:   handlers.NewPaymentHandler(paySvc, couponRepo.Validate),
		Branch:    handlers.NewBranchHandler(branchRepo, userRepo, authSvc),
		Store:     handlers.NewStoreHandler(storeRepo),
		PubKey:    pubKey,
		AdminIPs:  cfg.AdminIPs,
	}

	router.Setup(r, deps)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("servidor HTTP listo", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("fallo del servidor", "err", err)
			os.Exit(1)
		}
	}()

	// ── Graceful Shutdown ────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("iniciando apagado graceful...")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	if err := srv.Shutdown(ctx2); err != nil {
		slog.Error("shutdown forzado", "err", err)
	}
	slog.Info("servidor detenido")
}

func setupLogger(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})))
}
