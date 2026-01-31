# GoatSync + EteSync-DAV Self-Hosting Guide

> ⚠️ **NOTE**: This is a personal guide, not committed to git.

---

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Quick Start: Full Stack with Docker Compose](#quick-start-full-stack)
3. [Option 1: GoatSync Only](#option-1-goatsync-only)
4. [Option 2: GoatSync + etesync-dav](#option-2-goatsync--etesync-dav)
5. [Option 3: Binary Installation](#option-3-binary-installation)
6. [Connecting Clients](#connecting-clients)
7. [Reverse Proxy Setup](#reverse-proxy-setup)
8. [Backups](#backups)
9. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

### What Each Component Does

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           FULL ARCHITECTURE                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  YOUR DEVICES                                                                │
│  ════════════                                                                │
│                                                                              │
│  ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐       │
│  │  EteSync Apps   │     │   Thunderbird   │     │  Apple Calendar │       │
│  │ (iOS/Android/   │     │    Evolution    │     │   Outlook, etc  │       │
│  │     Web)        │     │    Contacts     │     │                 │       │
│  └────────┬────────┘     └────────┬────────┘     └────────┬────────┘       │
│           │                       │                       │                 │
│           │                       └───────────┬───────────┘                 │
│           │                                   │                             │
│           │ EteSync Protocol                  │ CalDAV/CardDAV              │
│           │ (MessagePack)                     │ (iCal/vCard)                │
│           │                                   │                             │
│           │                                   ▼                             │
│           │                       ┌─────────────────────┐                   │
│           │                       │   etesync-dav       │                   │
│           │                       │   (Python bridge)   │                   │
│           │                       │   Port: 37358       │                   │
│           │                       │                     │                   │
│           │                       │ • Translates CalDAV │                   │
│           │                       │   to EteSync API    │                   │
│           │                       │ • Decrypts data     │                   │
│           │                       │   for standard apps │                   │
│           │                       │ • Web UI at :37358  │                   │
│           │                       └──────────┬──────────┘                   │
│           │                                  │                              │
│           │                    EteSync Protocol                             │
│           │                                  │                              │
│           └──────────────────────────────────┤                              │
│                                              │                              │
│                                              ▼                              │
│                              ┌─────────────────────────┐                    │
│  YOUR SERVER                 │       GoatSync          │                    │
│  ═══════════                 │     (Go server)         │                    │
│                              │     Port: 3735          │                    │
│                              │                         │                    │
│                              │ • Stores encrypted data │                    │
│                              │ • User authentication   │                    │
│                              │ • Sync management       │                    │
│                              │ • WebSocket real-time   │                    │
│                              └───────────┬─────────────┘                    │
│                                          │                                  │
│                    ┌─────────────────────┼─────────────────────┐            │
│                    │                     │                     │            │
│           ┌────────▼────────┐   ┌────────▼────────┐   ┌───────▼───────┐   │
│           │   PostgreSQL    │   │     Redis       │   │ Chunk Storage │   │
│           │   (database)    │   │  (real-time)    │   │   (files)     │   │
│           └─────────────────┘   └─────────────────┘   └───────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### When Do You Need etesync-dav?

| Your Use Case | GoatSync Only | GoatSync + etesync-dav |
|---------------|---------------|------------------------|
| EteSync iOS app | ✅ | ✅ |
| EteSync Android app | ✅ | ✅ |
| EteSync Web app | ✅ | ✅ |
| Apple Calendar/Contacts | ❌ | ✅ |
| Thunderbird | ❌ | ✅ |
| Evolution/GNOME Calendar | ❌ | ✅ |
| Outlook | ❌ | ✅ |
| Any CalDAV/CardDAV app | ❌ | ✅ |

---

## Quick Start: Full Stack

### Complete docker-compose.yml (GoatSync + etesync-dav)

```yaml
version: '3.8'

services:
  # ═══════════════════════════════════════════════════════════
  # DATABASE
  # ═══════════════════════════════════════════════════════════
  postgres:
    image: postgres:15-alpine
    container_name: goatsync-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: goatsync
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
      POSTGRES_DB: goatsync
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goatsync"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ═══════════════════════════════════════════════════════════
  # REDIS (Optional - for WebSocket real-time sync)
  # ═══════════════════════════════════════════════════════════
  redis:
    image: redis:7-alpine
    container_name: goatsync-redis
    restart: unless-stopped
    volumes:
      - redis_data:/data

  # ═══════════════════════════════════════════════════════════
  # GOATSYNC SERVER (Main API server)
  # ═══════════════════════════════════════════════════════════
  goatsync:
    image: ghcr.io/YOUR_USERNAME/goatsync:latest
    # Or build locally: build: .
    container_name: goatsync
    restart: unless-stopped
    ports:
      - "3735:3735"
    environment:
      DATABASE_URL: postgres://goatsync:${DB_PASSWORD:-changeme}@postgres:5432/goatsync?sslmode=disable
      ENCRYPTION_SECRET: ${ENCRYPTION_SECRET:?ENCRYPTION_SECRET is required}
      REDIS_URL: redis://redis:6379/0
      PORT: "3735"
      DEBUG: "false"
    volumes:
      - chunk_data:/data/chunks
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started

  # ═══════════════════════════════════════════════════════════
  # ETESYNC-DAV (CalDAV/CardDAV bridge for standard apps)
  # ═══════════════════════════════════════════════════════════
  etesync-dav:
    image: etesync/etesync-dav:latest
    container_name: etesync-dav
    restart: unless-stopped
    ports:
      - "37358:37358"
    environment:
      # Point to GoatSync server (internal Docker network)
      ETESYNC_URL: http://goatsync:3735
    volumes:
      - dav_data:/data
    depends_on:
      - goatsync

volumes:
  postgres_data:
  redis_data:
  chunk_data:
  dav_data:
```

### Setup Steps

#### Step 1: Create .env file

```bash
mkdir -p ~/goatsync && cd ~/goatsync

# Create environment file
cat > .env << 'EOF'
# Database password
DB_PASSWORD=your_secure_database_password_here

# Encryption secret (min 32 characters)
# Generate with: openssl rand -base64 32
ENCRYPTION_SECRET=your_very_long_secret_key_at_least_32_characters
EOF

# Secure the file
chmod 600 .env
```

#### Step 2: Save docker-compose.yml

Save the compose file above to `~/goatsync/docker-compose.yml`

#### Step 3: Start Everything

```bash
cd ~/goatsync
docker compose up -d
```

#### Step 4: Verify Services

```bash
# Check all services are running
docker compose ps

# Check GoatSync health
curl http://localhost:3735/health
# {"status":"ok"}

# Check GoatSync API
curl http://localhost:3735/api/v1/authentication/is_etebase/
# (empty 200 OK)

# Check etesync-dav web UI
curl http://localhost:37358
# Should return HTML
```

#### Step 5: Create Your First Account

1. Open **EteSync Web App**: https://pim.etesync.com
2. Click **"Change Server"**
3. Enter your server: `http://YOUR_SERVER_IP:3735`
4. Click **"Sign Up"** and create account

#### Step 6: Set Up etesync-dav (for Thunderbird, Apple Calendar, etc.)

1. Open browser: `http://YOUR_SERVER_IP:37358`
2. Click **"Add Account"**
3. Enter:
   - Username: Your EteSync username
   - Server URL: `http://goatsync:3735` (or `http://YOUR_SERVER_IP:3735`)
   - Password: Your EteSync password
4. Click **"Copy Password"** - this is your DAV password!

---

## Option 1: GoatSync Only

If you only use EteSync native apps (no Thunderbird/Apple Calendar):

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: goatsync
      POSTGRES_PASSWORD: changeme
      POSTGRES_DB: goatsync
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goatsync"]
      interval: 10s
      timeout: 5s
      retries: 5

  goatsync:
    image: ghcr.io/YOUR_USERNAME/goatsync:latest
    restart: unless-stopped
    ports:
      - "3735:3735"
    environment:
      DATABASE_URL: postgres://goatsync:changeme@postgres:5432/goatsync?sslmode=disable
      ENCRYPTION_SECRET: your_secret_key_at_least_32_characters
      PORT: "3735"
    volumes:
      - chunk_data:/data/chunks
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  postgres_data:
  chunk_data:
```

---

## Option 2: GoatSync + etesync-dav

See the [Full Stack docker-compose.yml](#complete-docker-composeyml-goatsync--etesync-dav) above.

### How etesync-dav Works

```
1. You create account on GoatSync (via EteSync app)
   
2. You add that account to etesync-dav:
   └── Browser: http://localhost:37358
   └── Click "Add Account"
   └── Enter your EteSync credentials
   
3. etesync-dav generates a DAV password for you
   └── This is different from your EteSync password!
   └── Click "Copy Password" button
   
4. You use this DAV password in Thunderbird/Apple Calendar:
   └── Server: http://localhost:37358/your-username/
   └── Username: your-email@example.com
   └── Password: (the DAV password from step 3)
   
5. Data flow:
   Thunderbird → etesync-dav → GoatSync → PostgreSQL
                (translates)   (stores)
```

### etesync-dav Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ETESYNC_URL` | `https://api.etebase.com/` | GoatSync server URL |
| `ETESYNC_DATA_DIR` | `/data` | Where to store local cache |
| `ETESYNC_SERVER_HOSTS` | `0.0.0.0:37358` | Listen address |

---

## Option 3: Binary Installation

### Install GoatSync Binary

```bash
# Download
LATEST=$(curl -s https://api.github.com/repos/YOUR_USERNAME/goatsync/releases/latest | grep tag_name | cut -d '"' -f 4)
wget https://github.com/YOUR_USERNAME/goatsync/releases/download/${LATEST}/goatsync-linux-amd64
chmod +x goatsync-linux-amd64
sudo mv goatsync-linux-amd64 /usr/local/bin/goatsync
```

### Install etesync-dav

```bash
# Option A: Docker (recommended)
docker run -d \
  --name etesync-dav \
  -p 37358:37358 \
  -e ETESYNC_URL=http://localhost:3735 \
  -v etesync-dav:/data \
  --restart=always \
  etesync/etesync-dav:latest

# Option B: pip (requires Python 3.10+)
pip install etesync-dav
etesync-dav manage add YOUR_USERNAME
etesync-dav
```

### Systemd Services

**GoatSync service** (`/etc/systemd/system/goatsync.service`):
```ini
[Unit]
Description=GoatSync Server
After=network.target postgresql.service

[Service]
Type=simple
User=goatsync
ExecStart=/usr/local/bin/goatsync
Restart=always
Environment=DATABASE_URL=postgres://goatsync:password@localhost:5432/goatsync?sslmode=disable
Environment=ENCRYPTION_SECRET=your_secret_key_32_chars
Environment=PORT=3735

[Install]
WantedBy=multi-user.target
```

**etesync-dav service** (`/etc/systemd/system/etesync-dav.service`):
```ini
[Unit]
Description=EteSync DAV Bridge
After=network.target goatsync.service

[Service]
Type=simple
User=etesync
ExecStart=/usr/local/bin/etesync-dav
Restart=always
Environment=ETESYNC_URL=http://localhost:3735
Environment=ETESYNC_DATA_DIR=/var/lib/etesync-dav

[Install]
WantedBy=multi-user.target
```

---

## Connecting Clients

### EteSync Native Apps (Direct to GoatSync)

#### EteSync Web App
1. Go to https://pim.etesync.com
2. Click "Change Server"
3. Enter: `https://sync.yourdomain.com` (or `http://YOUR_IP:3735`)
4. Sign up or log in

#### EteSync Android
1. Download from Play Store/F-Droid
2. Tap "Advanced"
3. Enter server URL
4. Sign up or log in

#### EteSync iOS
1. Download from App Store
2. Tap "Custom Server"
3. Enter server URL
4. Sign up or log in

### CalDAV/CardDAV Apps (Through etesync-dav)

**First: Get your DAV password**
1. Open `http://YOUR_SERVER:37358`
2. Login with EteSync credentials
3. Click "Copy Password" - save this!

#### Thunderbird (91+)
1. **Calendar**: Click "+" in calendar view
2. Select "On the Network"
3. Username: your EteSync email
4. Location: `http://YOUR_SERVER:37358/your-email/calendars/`
5. Click "Find Calendars"
6. Enter DAV password when prompted

1. **Contacts**: File → New → CardDAV Address Book
2. Username: your EteSync email
3. Location: `http://YOUR_SERVER:37358/your-email/contacts/`
4. Enter DAV password

#### Apple Calendar (macOS)
1. System Preferences → Internet Accounts
2. Add Other Account → CalDAV
3. Account Type: Manual
4. Username: your EteSync email
5. Password: DAV password
6. Server: `YOUR_SERVER:37358`

#### Evolution / GNOME Calendar
1. File → New → Collection Account
2. Username: your EteSync email
3. Server: `http://YOUR_SERVER:37358/`
4. Check "Look up for a CalDAV/CardDAV server"
5. Enter DAV password when prompted

---

## Reverse Proxy Setup

### Caddy (Recommended - Auto HTTPS)

```
# /etc/caddy/Caddyfile

# GoatSync API
sync.yourdomain.com {
    reverse_proxy localhost:3735
}

# etesync-dav (CalDAV/CardDAV)
dav.yourdomain.com {
    reverse_proxy localhost:37358
}
```

### Nginx

```nginx
# GoatSync
server {
    listen 443 ssl http2;
    server_name sync.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/sync.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sync.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3735;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}

# etesync-dav
server {
    listen 443 ssl http2;
    server_name dav.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/dav.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dav.yourdomain.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:37358;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Backups

### Full Backup Script

```bash
#!/bin/bash
# /opt/goatsync/backup.sh

BACKUP_DIR=/opt/goatsync/backups
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

echo "Starting backup: $DATE"

# 1. Backup PostgreSQL
docker exec goatsync-db pg_dump -U goatsync goatsync | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# 2. Backup GoatSync chunks
docker cp goatsync:/data/chunks - | gzip > $BACKUP_DIR/chunks_$DATE.tar.gz

# 3. Backup etesync-dav data
docker cp etesync-dav:/data - | gzip > $BACKUP_DIR/dav_$DATE.tar.gz

# 4. Delete old backups (keep 30 days)
find $BACKUP_DIR -mtime +30 -delete

echo "Backup completed: $DATE"
```

### Restore

```bash
# Stop services
docker compose stop

# Restore database
gunzip -c backup_db_20240101.sql.gz | docker exec -i goatsync-db psql -U goatsync goatsync

# Restore chunks (optional)
docker cp backup_chunks.tar.gz goatsync:/tmp/
docker exec goatsync tar -xzf /tmp/backup_chunks.tar.gz -C /data/

# Start services
docker compose start
```

---

## Troubleshooting

### Check Service Status

```bash
# All services
docker compose ps
docker compose logs -f

# Individual
docker compose logs goatsync
docker compose logs etesync-dav
docker compose logs postgres
```

### Common Issues

#### "Connection refused" to GoatSync
```bash
# Check if running
docker compose ps goatsync

# Check logs
docker compose logs goatsync

# Test directly
curl http://localhost:3735/health
```

#### "Could not connect to server" in etesync-dav
```bash
# Inside Docker, use service name:
ETESYNC_URL=http://goatsync:3735

# Outside Docker, use IP:
ETESYNC_URL=http://localhost:3735
```

#### "Login failed" in etesync-dav
- Make sure you're using your **EteSync password**, not the DAV password
- DAV password is only for Thunderbird/Apple Calendar
- Create account first via EteSync app/web

#### CalDAV not finding calendars
- Make sure URL includes your username: `http://server:37358/user@example.com/`
- Use the DAV password, not EteSync password

---

## Quick Reference

| Service | Port | URL | Purpose |
|---------|------|-----|---------|
| GoatSync | 3735 | `http://localhost:3735` | Main API server |
| etesync-dav | 37358 | `http://localhost:37358` | CalDAV/CardDAV bridge |
| PostgreSQL | 5432 | Internal | Database |
| Redis | 6379 | Internal | Real-time sync |

| Command | Description |
|---------|-------------|
| `docker compose up -d` | Start all services |
| `docker compose down` | Stop all services |
| `docker compose logs -f` | View logs |
| `docker compose restart goatsync` | Restart GoatSync |
| `curl localhost:3735/health` | Check GoatSync health |

---

## Useful Links

- [EteSync Apps](https://www.etesync.com/get-apps/)
- [etesync-dav Docker Hub](https://hub.docker.com/r/etesync/etesync-dav)
- [etesync-dav GitHub](https://github.com/etesync/etesync-dav)
- [EteSync Protocol Docs](https://docs.etebase.com/)
