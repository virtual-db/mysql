# =============================================================================
# virtual-db/mysql — Dockerfile
#
# Produces a minimal release image containing only the vdb-mysql binary.
# All dependencies are resolved from their published module versions; no local
# source replacements or go.work files are needed.
#
# Runtime environment variables:
#   VDB_LISTEN_ADDR      — address:port the server listens on (default :3306)
#   VDB_DB_NAME          — database name exposed to connecting clients
#   VDB_SOURCE_DSN       — DSN for the upstream MySQL source
#   VDB_AUTH_SOURCE_ADDR — host:port of the auth source MySQL instance
#   VDB_PLUGIN_DIR       — optional: directory containing plugin subdirectories
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: builder
# -----------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy only the module files first so Docker can cache the download layer
# independently of source changes.
COPY go.mod go.sum ./

RUN go mod download

# Copy source and build the binary.
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/vdb-mysql .

# -----------------------------------------------------------------------------
# Stage 2: runtime
# -----------------------------------------------------------------------------
FROM alpine:3.20

# ca-certificates — required for any TLS connections the binary makes at runtime.
# tzdata         — ensures time-zone-aware SQL functions behave correctly.
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/vdb-mysql /usr/local/bin/vdb-mysql

EXPOSE 3306

ENTRYPOINT ["/usr/local/bin/vdb-mysql"]
