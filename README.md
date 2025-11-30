# 🐐 GoatSync

GoatSync is a **Go implementation of the EteSync server** with 100% API compatibility.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)](https://www.postgresql.org/)

> **Status: ✅ Production Ready (v1.0.0)**
>
> Fully compatible with all EteSync clients (web, iOS, Android, etesync-dav).

![GoatSync](./screenshot.png)

## ✨ Features

- **🔐 End-to-end encryption** - Same security as original EteSync
- **📱 100% Client Compatible** - Works with all existing EteSync apps
- **⚡ High Performance** - Built with Go + Gin for maximum throughput
- **🐘 PostgreSQL** - Production-grade database with GORM
- **🔄 Real-time Sync** - WebSocket support with Redis pub/sub
- **🐳 Docker Ready** - One-command deployment

## 🚀 Quick Start

### 1. Start Dependencies

```bash
docker compose up -d
```

### 2. Build & Run

```bash
go build -o goatsync ./cmd/server

DATABASE_URL="postgres://goatsync:goatsync@localhost:5432/goatsync?sslmode=disable" \
SECRET_KEY="your-secret-key-at-least-32-characters" \
./goatsync
```

### 3. Verify

```bash
curl http://localhost:3735/health
# {"status":"ok"}
```

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [RUNNING.md](RUNNING.md) | Complete guide to running GoatSync |
| [CHANGELOG.md](CHANGELOG.md) | Version history and features |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Technical architecture details |
| [LLM_START_HERE.md](LLM_START_HERE.md) | Quick reference for AI assistants |

## ⚙️ Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `SECRET_KEY` | Yes | - | Encryption key (min 32 chars) |
| `PORT` | No | `8080` | HTTP server port |
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

## 🐳 Docker Deployment

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: goatsync
      POSTGRES_PASSWORD: goatsync
      POSTGRES_DB: goatsync
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine

  goatsync:
    build: .
    ports:
      - "3735:3735"
    environment:
      DATABASE_URL: postgres://goatsync:goatsync@postgres:5432/goatsync?sslmode=disable
      SECRET_KEY: your-secret-key-at-least-32-characters
      REDIS_URL: redis://redis:6379/0
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:
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
