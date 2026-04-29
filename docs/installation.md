# Installation

---

## Pre-built Binary (linux/amd64)

Linux/amd64 binaries are attached to every tagged release.

```sh
curl -Lo vdb-mysql https://github.com/virtual-db/mysql/releases/latest/download/vdb-mysql-linux-amd64
chmod +x vdb-mysql
./vdb-mysql
```

To pin to a specific version, replace `latest/download` with `download/v0.x.x`.

---

## Docker

An official image is not yet published. Build one locally from the included `Dockerfile`.

### Build from the repository

```sh
git clone https://github.com/virtual-db/mysql
cd mysql
docker build -t vdb-mysql .
```

### Build without cloning

```dockerfile
FROM golang:1.23-alpine AS builder
RUN CGO_ENABLED=0 go install github.com/virtual-db/mysql@latest

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /go/bin/mysql /usr/local/bin/vdb-mysql
EXPOSE 3306
ENTRYPOINT ["/usr/local/bin/vdb-mysql"]
```

```sh
docker build -t vdb-mysql -f Dockerfile.custom .
```

### Docker Compose

```yaml
services:
  vdb:
    image: vdb-mysql
    restart: unless-stopped
    environment:
      VDB_LISTEN_ADDR: ":3306"
      VDB_DB_NAME: myapp
      VDB_SOURCE_DSN: "vdb_user:secret@tcp(db.internal:3306)/myapp"
      VDB_AUTH_SOURCE_ADDR: "db.internal:3306"
    ports:
      - "3306:3306"
```

```sh
docker compose up -d
```

---

## Build From Source

Requires Go 1.23.3 or later. All dependencies are published to the public Go
module proxy — no private module access is required.

```sh
git clone https://github.com/virtual-db/mysql
cd mysql
CGO_ENABLED=0 go build -trimpath -o vdb-mysql .
./vdb-mysql
```

To cross-compile for a different target:

```sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o vdb-mysql-linux-amd64 .
```

---

## Requirements

| Dependency | Version |
|---|---|
| Source database | MySQL 8.x |
| Go (build from source only) | 1.23.3 or later |
| Docker (container deployment) | 20.10 or later |
| Network | vdb-mysql must have TCP access to the source MySQL server on the port in `VDB_SOURCE_DSN` and `VDB_AUTH_SOURCE_ADDR` |

---

## Next Steps

Once vdb-mysql is running, see:

- [Configuration](./configuration.md) — environment variables and source database setup
- [How It Works](./how-it-works.md) — architecture overview
- [Plugins](./plugins.md) — extending vdb-mysql with plugins