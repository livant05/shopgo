#!/usr/bin/env bash
set -euo pipefail

echo "🔧 ShopGo — Setup inicial"

# Dependencias
command -v docker  >/dev/null 2>&1 || { echo "❌ Docker requerido"; exit 1; }
command -v go      >/dev/null 2>&1 || { echo "❌ Go 1.22+ requerido"; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "❌ OpenSSL requerido"; exit 1; }

# .env
if [ ! -f .env ]; then
  cp .env.example .env
  echo "✅ .env creado — edita STRIPE_SECRET_KEY y STRIPE_WEBHOOK_SECRET"
fi

# JWT keys
mkdir -p deploy/secrets
if [ ! -f deploy/secrets/jwt_private.pem ]; then
  openssl genrsa -out deploy/secrets/jwt_private.pem 2048
  openssl rsa -in deploy/secrets/jwt_private.pem -pubout -out deploy/secrets/jwt_public.pem
  chmod 600 deploy/secrets/jwt_private.pem
  echo "✅ Claves JWT RSA 2048-bit generadas"
fi

# Docker
docker compose up -d postgres redis minio
echo "⏳ Esperando PostgreSQL..."
until docker compose exec -T postgres pg_isready -U shopuser -d shopgo >/dev/null 2>&1; do sleep 1; done

# Migraciones
make migrate
echo "✅ Migraciones aplicadas"

# MinIO
docker compose exec -T minio mc alias set local http://localhost:9000 minioadmin minioadmin 2>/dev/null || true
docker compose exec -T minio mc mb --ignore-existing local/shopgo 2>/dev/null || true
docker compose exec -T minio mc anonymous set public local/shopgo/public 2>/dev/null || true
echo "✅ MinIO configurado"

echo ""
echo "🚀 Listo! Comandos:"
echo "   make dev        → backend :8080"
echo "   make ui         → admin :5174 + storefront :5173"
echo "   http://localhost:9001  → MinIO (minioadmin/minioadmin)"
