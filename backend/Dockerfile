# syntax=docker/dockerfile:1

# ---------- Builder ----------
FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG BUILD_TIME=unknown

WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Build (with version stamping via ldflags)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags "-s -w \
      -X github.com/aeroxe/approval-flow/internal/config.Version=${VERSION} \
      -X github.com/aeroxe/approval-flow/internal/config.BuildTime=${BUILD_TIME}" \
    -o /out/server ./cmd/server

# ---------- Runtime ----------
FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

# The server binary AND the versioned SQL migrations — the app relies on
# golang-migrate for schema management, never on AutoMigrate, in production.
COPY --from=builder /out/server /app/server
COPY --from=builder /src/migrations /app/migrations

# Run as an unprivileged user
USER app

ENV ENV=production \
    LOG_FILE_PATH="" \
    MIGRATIONS_PATH=/app/migrations

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health/live >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/server"]
