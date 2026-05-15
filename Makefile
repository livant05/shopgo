.PHONY: setup dev build test lint keys migrate sqlc ui

setup:
	@cp -n .env.example .env || true
	@$(MAKE) keys
	@docker compose up -d postgres redis minio
	@sleep 4
	@$(MAKE) migrate
	@echo "✅ Setup completado. Edita .env con tus claves de Stripe."

dev:
	@docker compose up -d postgres redis minio
	@air -c .air.toml

ui:
	@cd frontend && pnpm dev:all

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-s -w" -o bin/shopgo ./cmd/api

test:
	go test -race -timeout 60s ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "→ coverage.html generado"

lint:
	golangci-lint run ./...

security:
	gosec ./...

keys:
	@mkdir -p deploy/secrets
	@openssl genrsa -out deploy/secrets/jwt_private.pem 2048
	@openssl rsa -in deploy/secrets/jwt_private.pem -pubout -out deploy/secrets/jwt_public.pem
	@chmod 600 deploy/secrets/jwt_private.pem
	@echo "✅ Claves JWT generadas"

migrate:
	migrate -path db/migrations -database "$$(grep DATABASE_URL .env | cut -d= -f2-)" up

migrate-down:
	migrate -path db/migrations -database "$$(grep DATABASE_URL .env | cut -d= -f2-)" down 1

migrate-new:
	@read -p "Nombre: " name; migrate create -ext sql -dir db/migrations -seq $$name

sqlc:
	sqlc generate

prod-up:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

clean:
	@rm -f bin/shopgo coverage.out coverage.html
