# 🐐 GoatSync

A **Go implementation of the EteSync server** with 100% API compatibility.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ghcr.io-blue?logo=docker)](https://github.com/jollySleeper/GoatSync/pkgs/container/goatsync)

> Fully compatible with all EteSync clients (web, iOS, Android, etesync-dav).

## ✨ Features

- **🔐 End-to-end encryption** - Same security as original EteSync
- **📱 100% Client Compatible** - Works with all existing EteSync apps
- **⚡ High Performance** - Built with Go + Gin for maximum throughput
- **🐘 PostgreSQL** - Production-grade database with GORM
- **🔄 Real-time Sync** - WebSocket support with Redis pub/sub
- **🐳 Docker Ready** - One-command deployment with multi-arch support

## 🚀 Quick Start

```bash
# Clone and configure
git clone https://github.com/jollySleeper/GoatSync.git
cd GoatSync
cp .env.example .env
# Edit .env: set ENCRYPTION_SECRET and DATABASE_URL password

# Start services
docker compose up -d

# Verify
curl http://localhost:3735/health
# {"status":"ok"}
```

For CalDAV/CardDAV support (Thunderbird, Apple Calendar, etc.):

```bash
docker compose -f docker-compose-full.yml up -d
```

## ⚙️ Configuration

See [.env.example](.env.example) for all options.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `ENCRYPTION_SECRET` | Yes | - | Encryption key (min 32 chars) |
| `PORT` | No | `3735` | Server port |
| `REDIS_URL` | No | - | Redis for WebSocket sync |
| `DEBUG` | No | `false` | Debug mode |

## 🐳 Docker Images

```bash
docker pull ghcr.io/jollysleeper/goatsync:latest
docker pull ghcr.io/jollysleeper/goatsync:0.1.2
```

Platforms: `linux/amd64`, `linux/arm64`

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Production deployment guide |
| [docs/RUNNING.md](docs/RUNNING.md) | Running locally |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Technical details |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | Version history |

## 🔌 API Endpoints

Implements all EteSync API endpoints:
- Authentication (signup, login, logout)
- Collections (CRUD with stoken pagination)
- Items (batch, transaction, revisions)
- Members & Invitations (sharing)
- Chunks (binary data)
- WebSocket (real-time sync)

## 🔒 Security

Same cryptographic protocols as EteSync:
- **BLAKE2b-256** - Key derivation
- **XSalsa20-Poly1305** - NaCl SecretBox encryption
- **Ed25519** - Signature verification

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [EteSync](https://github.com/etesync) - Original Python implementation
- [Gin](https://gin-gonic.com/) - HTTP framework
- [GORM](https://gorm.io/) - ORM library
