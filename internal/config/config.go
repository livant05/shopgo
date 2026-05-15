package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv    string
	Port      string
	StoreName string
	LogLevel  string

	DatabaseURL string
	DBMaxConns  int32
	DBMinConns  int32

	RedisAddr     string
	RedisPassword string

	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration

	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string

	S3Endpoint  string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string

	ResendAPIKey    string
	EmailFrom       string
	StoreAdminEmail string

	AllowedOrigins []string
	AdminIPs       []string

	DefaultTaxRate  float64
	DefaultCurrency string
}

func Load() (*Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}
	taxRate, _ := strconv.ParseFloat(getEnv("DEFAULT_TAX_RATE", "0.16"), 64)
	maxConns, _ := strconv.ParseInt(getEnv("DB_MAX_CONNS", "25"), 10, 32)
	minConns, _ := strconv.ParseInt(getEnv("DB_MIN_CONNS", "5"), 10, 32)

	return &Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "8080"),
		StoreName: getEnv("STORE_NAME", "Mi Tienda"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),

		DatabaseURL: must("DATABASE_URL"),
		DBMaxConns:  int32(maxConns),
		DBMinConns:  int32(minConns),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTPrivateKeyPath: must("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeyPath:  must("JWT_PUBLIC_KEY_PATH"),
		JWTAccessTTL:      accessTTL,
		JWTRefreshTTL:     refreshTTL,

		StripeSecretKey:      must("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:  must("STRIPE_WEBHOOK_SECRET"),
		StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),

		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
		S3Bucket:    getEnv("S3_BUCKET", "shopgo"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),

		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		EmailFrom:       getEnv("EMAIL_FROM", "no-reply@tienda.com"),
		StoreAdminEmail: getEnv("STORE_ADMIN_EMAIL", ""),

		AllowedOrigins: split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174")),
		AdminIPs:       split(getEnv("ADMIN_IPS", "")),

		DefaultTaxRate:  taxRate,
		DefaultCurrency: getEnv("DEFAULT_CURRENCY", "MXN"),
	}, nil
}

func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("variable de entorno requerida no definida: " + key)
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func split(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
