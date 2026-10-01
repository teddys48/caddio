# ============================================================
# Stage 1: Build Frontend
# ============================================================
FROM oven/bun:1-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/bun.lock* ./
RUN bun install --frozen-lockfile || bun install

COPY frontend/ ./

RUN bun run check
RUN bun run build


# ============================================================
# Stage 2: Build Backend
# ============================================================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app/backend

RUN apk add --no-cache git gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Frontend static files
COPY --from=frontend-builder /app/frontend/dist ./frontend_dist

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-s -w" \
    -o /app/caddy-manager \
    ./cmd/server


# ============================================================
# Stage 3: Production
# ============================================================
FROM alpine:3.21 AS runner

RUN apk add --no-cache \
    ca-certificates \
    tzdata

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

# Backend binary
COPY --from=backend-builder \
    /app/caddy-manager \
    ./caddy-manager

# IMPORTANT:
# Copy frontend dist into final image too.
COPY --from=frontend-builder \
    /app/frontend/dist \
    ./frontend_dist

# SQLite data
RUN mkdir -p /data \
    && chown -R appuser:appgroup /app /data

USER appuser

# ENV PORT=8080
# ENV DATABASE_PATH=/data/caddy-manager.db
# ENV CADDY_ADMIN_URL=http://127.0.0.1:2019
# ENV APP_ENV=production
# ENV TZ=Asia/Jakarta

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["./caddy-manager"]