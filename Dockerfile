# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .

RUN go build -mod=vendor -ldflags="-s -w" -o /shopgo ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /shopgo /app/shopgo
RUN addgroup -S shopgo && adduser -S shopgo -G shopgo
USER shopgo
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/shopgo"]
