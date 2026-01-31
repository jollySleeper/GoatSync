# 🐐 GoatSync

GoatSync is a **Go implementation of the EteSync server** with 100% API compatibility.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ghcr.io-blue?logo=docker)](https://github.com/jollySleeper/GoatSync/pkgs/container/goatsync)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)](https://www.postgresql.org/)

> **Status: ✅ Production Ready (v0.1.0)**
>
> Fully compatible with all EteSync clients (web, iOS, Android, etesync-dav).

## ✨ Features

- **🔐 End-to-end encryption** - Same security as original EteSync
- **📱 100% Client Compatible** - Works with all existing EteSync apps
- **⚡ High Performance** - Built with Go + Gin for maximum throughput
- **🐘 PostgreSQL** - Production-grade database with GORM
- **🔄 Real-time Sync** - WebSocket support with Redis pub/sub
- **🐳 Docker Ready** - One-command deployment with multi-arch support

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone repository
git clone https://github.com/jollySleeper/GoatSync.git
cd GoatSync

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start all services
docker compose up -d

# Verify
curl http://localhost:3735/health
# {"status":"ok"}
```

### Option 2: Build from Source

```bash
# Start dependencies
docker compose up -d postgres redis

# Build and run
go build -o goatsync ./cmd/server
DATABASE_URL="postgres://goatsync:goatsync@localhost:5432/goatsync?sslmode=disable" \
ENCRYPTION_SECRET="your-secret-key-at-least-32-characters" \
./goatsync
```

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Complete deployment guide with Docker |
| [docs/RUNNING.md](docs/RUNNING.md) | Running GoatSync locally |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | Version history and features |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Technical architecture details |
| [docs/COMPARISON.md](docs/COMPARISON.md) | EteSync vs GoatSync comparison |

## ⚙️ Configuration

See [.env.example](.env.example) for all configuration options.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `ENCRYPTION_SECRET` | Yes | - | Encryption key (min 32 chars) |
| `PORT` | No | `3735` | HTTP server port |
| `REDIS_URL` | No | - | Redis for WebSocket pub/sub |
| `DEBUG` | No | `false` | Enable debug mode |

## 🔌 API Endpoints

GoatSync implements all EteSync API endpoints:

- **Authentication** - Signup, login, logout, password change
- **Collections** - CRUD operations with stoken pagination
- **Items** - Batch, transaction, fetch updates, revisions
- **Members** - Sharing and access control
- **Invitations** - Incoming/outgoing invitation management
- **Chunks** - Binary data upload/download
- **WebSocket** - Real-time sync notifications

See [RUNNING.md](RUNNING.md) for the complete API reference.

## 🧪 Testing

```bash
# Unit tests
go test ./... -v

# Integration tests (requires Docker)
docker compose up -d
go test ./internal/integration/... -v
```

## 🐳 Docker

### Docker Images

Official images are published to GitHub Container Registry:

```bash
# Latest stable release
docker pull ghcr.io/jollysleeper/goatsync:latest

# Specific version
docker pull ghcr.io/jollysleeper/goatsync:0.1.0
```

**Supported platforms:** `linux/amd64`, `linux/arm64`

### Docker Compose

Two compose files are provided:

| File | Description |
|------|-------------|
| `docker-compose.yml` | GoatSync + PostgreSQL + Redis |
| `docker-compose-full.yml` | Above + EteSync-DAV for CalDAV/CardDAV |

```bash
# Standard deployment
docker compose up -d

# With CalDAV/CardDAV support (for Thunderbird, Apple Calendar, etc.)
docker compose -f docker-compose-full.yml up -d
```

## 🏗️ Architecture

```
goatSync/
├── cmd/server/          # Entry point
├── internal/
│   ├── crypto/          # BLAKE2b, SecretBox, Ed25519
│   ├── database/        # GORM PostgreSQL
│   ├── model/           # 9 GORM models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Auth, CORS
│   └── server/          # HTTP server
└── pkg/errors/          # EtebaseError types
```

## 🔐 Security

GoatSync implements the same cryptographic protocols as EteSync:

- **BLAKE2b-256** - Key derivation with salt and personalization
- **XSalsa20-Poly1305** - NaCl SecretBox for symmetric encryption
- **Ed25519** - Signature verification for authentication

**⚠️ Never use bcrypt** - EteSync uses Ed25519 signatures, not password hashing.

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [EteSync](https://github.com/etesync) - The original Python implementation
- [Gin](https://gin-gonic.com/) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library

---

**Made with ❤️ by the GoatSync community**
