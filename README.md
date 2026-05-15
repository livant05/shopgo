# ShopGo — E-Commerce Profesional

> "Complexity is the enemy of reliability. Build what you need, not what you might need."
> — 20 years of scars speaking

Sistema de e-commerce **single-tenant** construido con integridad arquitectónica.
Sin multi-empresa, sin multi-schema, sin complejidad innecesaria.

## Stack

| Capa | Tecnología | Por qué |
|------|-----------|---------|
| API | Go 1.22 + Gin | Rendimiento, tipado, tooling maduro |
| BD | PostgreSQL 16 | Un schema, transacciones ACID reales |
| Caché | Redis 7 | Cache + Pub/Sub WebSockets |
| Pagos | Stripe (directo) | Sin Connect overhead — 0% comisión de plataforma |
| Frontend Admin | Vue 3 + Pinia | Reactivo, liviano, composición clara |
| Frontend Tienda | Vue 3 + Pinia | SSR-ready con Vite |
| Infra | Docker + K8s | Deploy reproducible |

## Decisiones Arquitectónicas

### ¿Por qué single-tenant?
Si tu negocio es una sola empresa, **schema-per-tenant es complejidad que no necesitas**.
TenantPool con RWMutex, X-Tenant-Slug headers, search_path dinámico — todo eso es
cognitive overhead que pagas cada día en mantenimiento. Un pool de conexiones,
un schema, consultas directas.

### ¿Por qué Stripe directo en lugar de Connect?
Stripe Connect es para marketplaces. Si TÚ eres el comercio, usas Stripe directo:
sin ApplicationFee, sin Transfer, sin el 2-3% extra de plataforma.
Tu dinero va directo a tu cuenta Stripe.

### Branches (Sucursales) sin Company
El negocio puede tener múltiples sucursales. Pero no necesita una capa Company encima.
La sucursal es la unidad de operación. El sistema entero sirve a una sola empresa.

## Inicio Rápido

```bash
make setup    # claves JWT + Docker + migraciones
make dev      # backend :8080
make ui       # admin :5174 + storefront :5173
```

## Comandos

```bash
make test         # tests con race detector
make lint         # golangci-lint + gosec
make migrate      # aplicar migraciones
make sqlc         # regenerar queries type-safe
make build        # binario producción
make prod-up      # docker-compose producción
```
