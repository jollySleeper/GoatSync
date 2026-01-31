# GoatSync Deployment Guide

A complete guide to deploying GoatSync in production using Docker.

## Prerequisites

- Docker Engine 20.10+ and Docker Compose v2+
- A server with at least 1GB RAM (2GB recommended)
- Domain name (optional, for HTTPS)

## Quick Start

### 1. Clone or Download

```bash
# Option A: Clone the repository
git clone https://github.com/jollySleeper/GoatSync.git
cd GoatSync

# Option B: Just download the docker-compose files
mkdir goatsync && cd goatsync
curl -O https://raw.githubusercontent.com/jollySleeper/GoatSync/main/docker-compose.yml
curl -O https://raw.githubusercontent.com/jollySleeper/GoatSync/main/.env.example
```

### 2. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Generate secure passwords
echo "DB_PASSWORD=$(openssl rand -base64 16)" >> .env
echo "SECRET_KEY=$(openssl rand -base64 32)" >> .env

# Edit to customize (or leave defaults)
nano .env

# Secure the file
chmod 600 .env
```

### 3. Start Services

```bash
# GoatSync only (for EteSync mobile/web apps)
docker compose up -d

# Full stack with CalDAV/CardDAV support (for Thunderbird, Apple Calendar, etc.)
docker compose -f docker-compose-full.yml up -d
```

### 4. Verify Deployment

```bash
# Check all services are healthy
docker compose ps

# Test GoatSync API
curl http://localhost:3735/health
# Expected: {"status":"ok"}

curl http://localhost:3735/api/v1/authentication/is_etebase/
# Expected: 200 OK (empty body)
```

## Docker Compose Files

### `docker-compose.yml` - Standard Deployment

Includes:
- **PostgreSQL** - Database for user accounts and metadata
- **Redis** - Real-time WebSocket synchronization
- **GoatSync** - Main EteSync-compatible API server

Use this if you only need EteSync mobile/web apps.

### `docker-compose-full.yml` - Full Stack Deployment

Includes everything above, plus:
- **EteSync-DAV** - CalDAV/CardDAV bridge for standard calendar apps

Use this if you want to use Thunderbird, Apple Calendar, Evolution, or any CalDAV/CardDAV client.

## Configuration Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | PostgreSQL password | `MySecurePass123!` |
| `SECRET_KEY` | Encryption secret (min 32 chars) | `$(openssl rand -base64 32)` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DEBUG` | `false` | Enable debug logging |
| `PORT` | `3735` | GoatSync server port |
| `GIN_MODE` | `release` | Gin framework mode |
| `CHALLENGE_VALID_SECONDS` | `300` | Login challenge validity |
| `ALLOWED_ORIGINS` | `*` | CORS origins (comma-separated) |
| `ALLOWED_HOSTS` | `*` | Allowed Host headers |

## Docker Images

Official images are published to GitHub Container Registry:

```bash
# Latest stable release
docker pull ghcr.io/jollysleeper/goatsync:latest

# Specific version
docker pull ghcr.io/jollysleeper/goatsync:0.1.0

# Development (main branch)
docker pull ghcr.io/jollysleeper/goatsync:main
```

### Available Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release (recommended) |
| `0.1.0`, `0.1`, etc. | Specific version |
| `main` | Latest development build |
| `sha-xxxxxx` | Specific commit |

### Supported Platforms

- `linux/amd64` - Standard x86_64 servers
- `linux/arm64` - ARM servers (Raspberry Pi 4, Apple Silicon, AWS Graviton)

## Production Deployment

### Reverse Proxy with HTTPS (Recommended)

For production, use a reverse proxy like Caddy, nginx, or Traefik.

**Caddy (easiest):**

```Caddyfile
# /etc/caddy/Caddyfile
sync.yourdomain.com {
    reverse_proxy localhost:3735
}

dav.yourdomain.com {
    reverse_proxy localhost:37358
}
```

**nginx:**

```nginx
server {
    listen 443 ssl http2;
    server_name sync.yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://127.0.0.1:3735;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Security Hardening

1. **Use strong passwords**: Generate with `openssl rand -base64 32`
2. **Enable HTTPS**: Never run in production without TLS
3. **Firewall**: Only expose ports 80/443 through reverse proxy
4. **Regular backups**: See backup section below

### Resource Limits

Add resource limits for production stability:

```yaml
services:
  goatsync:
    image: ghcr.io/jollysleeper/goatsync:latest
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1'
        reservations:
          memory: 128M
```

## Backups

### Database Backup

```bash
# Create backup
docker exec goatsync-postgres pg_dump -U goatsync goatsync > backup_$(date +%Y%m%d).sql

# Restore backup
cat backup_20260131.sql | docker exec -i goatsync-postgres psql -U goatsync goatsync
```

### Full Data Backup

```bash
# Stop services
docker compose down

# Backup all volumes
docker run --rm -v goatsync_postgres_data:/data -v $(pwd):/backup alpine \
    tar czf /backup/postgres_backup.tar.gz /data

docker run --rm -v goatsync_chunk_data:/data -v $(pwd):/backup alpine \
    tar czf /backup/chunks_backup.tar.gz /data

# Restart services
docker compose up -d
```

### Automated Backups

Create a cron job for daily backups:

```bash
# /etc/cron.daily/goatsync-backup
#!/bin/bash
BACKUP_DIR=/var/backups/goatsync
mkdir -p $BACKUP_DIR
docker exec goatsync-postgres pg_dump -U goatsync goatsync | gzip > $BACKUP_DIR/db_$(date +%Y%m%d).sql.gz
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete  # Keep 7 days
```

## Upgrading

### Standard Upgrade

```bash
cd /path/to/goatsync

# Pull latest images
docker compose pull

# Restart with new images
docker compose up -d

# Verify health
docker compose ps
curl http://localhost:3735/health
```

### Upgrade to Specific Version

```bash
# Edit docker-compose.yml to use specific version
# image: ghcr.io/jollysleeper/goatsync:0.2.0

docker compose up -d
```

## Troubleshooting

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f goatsync

# Last 100 lines
docker compose logs --tail=100 goatsync
```

### Common Issues

**Database connection failed:**
```bash
# Check if postgres is healthy
docker compose ps
docker compose logs postgres

# Reset database (WARNING: deletes all data)
docker compose down -v
docker compose up -d
```

**Port already in use:**
```bash
# Check what's using the port
lsof -i :3735

# Change port in docker-compose.yml
ports:
  - "8080:3735"  # Use port 8080 instead
```

**Out of disk space:**
```bash
# Clean up Docker
docker system prune -a

# Check volume sizes
docker system df -v
```

## Connecting Clients

### EteSync Mobile/Web Apps

1. Open EteSync app (iOS, Android, or https://pim.etesync.com)
2. Click "Change Server" or "Custom Server"
3. Enter: `https://sync.yourdomain.com` (or `http://YOUR_IP:3735`)
4. Sign up or log in

### CalDAV/CardDAV Apps (requires docker-compose-full.yml)

1. Open `https://dav.yourdomain.com` (or `http://YOUR_IP:37358`)
2. Add your EteSync account
3. Copy the generated DAV password
4. Configure your calendar app:
   - **CalDAV URL**: `https://dav.yourdomain.com/YOUR_USERNAME/calendars/`
   - **CardDAV URL**: `https://dav.yourdomain.com/YOUR_USERNAME/contacts/`
   - **Username**: Your EteSync username
   - **Password**: The DAV password (not your EteSync password!)

## Support

- **Issues**: https://github.com/jollySleeper/GoatSync/issues
- **Discussions**: https://github.com/jollySleeper/GoatSync/discussions
