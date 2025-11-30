# Running GoatSync

This guide explains how to run GoatSync and verify compatibility with the original EteSync server.

## Quick Start

### 1. Start Dependencies (PostgreSQL + Redis)

```bash
cd goatSync
docker compose up -d
```

This starts:
- **PostgreSQL** on `localhost:5432` (user: goatsync, pass: goatsync, db: goatsync)
- **Redis** on `localhost:6379`

### 2. Build GoatSync

```bash
go build -o goatsync ./cmd/server
```

### 3. Run GoatSync

```bash
# Set environment variables
export DATABASE_URL="postgres://goatsync:goatsync@localhost:5432/goatsync?sslmode=disable"
export SECRET_KEY="your-secret-key-at-least-32-characters-long"
export REDIS_URL="redis://localhost:6379/0"
export DEBUG=true
export PORT=3735

# Run the server
./goatsync
```

### 4. Verify It's Running

```bash
# Health check
curl http://localhost:3735/health
# {"status":"ok"}

# EteSync check
curl http://localhost:3735/api/v1/authentication/is_etebase/
# (empty 200 OK response)
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `SECRET_KEY` | Yes | - | Encryption key (min 32 chars) |
| `REDIS_URL` | No | - | Redis URL for WebSocket pub/sub |
| `PORT` | No | `8080` | HTTP server port |
| `DEBUG` | No | `false` | Enable debug mode |
| `CHUNK_STORAGE_PATH` | No | `/tmp/goatsync/chunks` | Chunk file storage path |
| `ALLOWED_ORIGINS` | No | `*` | CORS allowed origins (comma-separated) |

---

## Running with Docker Compose (Full Stack)

Create a `docker-compose.full.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: goatsync
      POSTGRES_PASSWORD: goatsync
      POSTGRES_DB: goatsync
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U goatsync"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  goatsync:
    build: .
    ports:
      - "3735:3735"
    environment:
      DATABASE_URL: postgres://goatsync:goatsync@postgres:5432/goatsync?sslmode=disable
      SECRET_KEY: your-secret-key-at-least-32-characters-long
      REDIS_URL: redis://redis:6379/0
      PORT: "3735"
      DEBUG: "true"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started

volumes:
  postgres_data:
  redis_data:
```

Then run:
```bash
docker compose -f docker-compose.full.yml up -d
```

---

## Running Tests

### Unit Tests
```bash
go test ./... -v
```

### Integration Tests (requires running database)
```bash
# Start dependencies first
docker compose up -d

# Run integration tests
go test ./internal/integration/... -v
```

### Run with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Verifying 1:1 Compatibility with EteSync Server

### Method 1: Side-by-Side Testing

1. **Start the Python EteSync server:**
```bash
cd eteSync-server
pip install -r requirements.txt
python manage.py migrate
python manage.py runserver 8000
```

2. **Start GoatSync:**
```bash
cd goatSync
docker compose up -d
./goatsync  # Runs on port 3735
```

3. **Compare responses:**
```bash
# Test is_etebase endpoint
curl -s http://localhost:8000/api/v1/authentication/is_etebase/ | xxd
curl -s http://localhost:3735/api/v1/authentication/is_etebase/ | xxd

# Test login_challenge (with msgpack body)
# Use a tool like httpie with msgpack support
```

### Method 2: Use Real EteSync Clients

Configure EteSync clients to point to GoatSync:

1. **etesync-dav:**
```bash
etesync-dav --server-url http://localhost:3735
```

2. **EteSync Web App:**
   - Modify the server URL in settings
   - Or use browser dev tools to redirect API calls

3. **Mobile Apps:**
   - Use a proxy to redirect to GoatSync

### Method 3: Integration Test Suite

The integration tests in `internal/integration/integration_test.go` verify:
- Health endpoints
- Signup flow
- Login challenge flow
- Unauthorized access protection
- Crypto compatibility
- MessagePack serialization
- Error response format
- Stoken generation
- Ed25519 signature verification

Run them:
```bash
go test ./internal/integration/... -v
```

---

## API Endpoints Reference

### Authentication
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/authentication/is_etebase/` | No |
| POST | `/api/v1/authentication/signup/` | No |
| POST | `/api/v1/authentication/login_challenge/` | No |
| POST | `/api/v1/authentication/login/` | No |
| POST | `/api/v1/authentication/logout/` | Yes |
| POST | `/api/v1/authentication/change_password/` | Yes |
| POST | `/api/v1/authentication/dashboard_url/` | Yes |

### Collections
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/collection/` | Yes |
| POST | `/api/v1/collection/` | Yes |
| POST | `/api/v1/collection/list_multi/` | Yes |
| GET | `/api/v1/collection/:uid/` | Yes |

### Items
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/collection/:uid/item/` | Yes |
| GET | `/api/v1/collection/:uid/item/:item_uid/` | Yes |
| GET | `/api/v1/collection/:uid/item/:item_uid/revision/` | Yes |
| POST | `/api/v1/collection/:uid/item/batch/` | Yes |
| POST | `/api/v1/collection/:uid/item/transaction/` | Yes |
| POST | `/api/v1/collection/:uid/item/fetch_updates/` | Yes |

### Chunks
| Method | Path | Auth Required |
|--------|------|---------------|
| PUT | `/api/v1/collection/:uid/item/:item_uid/chunk/:chunk_uid/` | Yes |
| GET | `/api/v1/collection/:uid/item/:item_uid/chunk/:chunk_uid/download/` | Yes |

### Members
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/collection/:uid/member/` | Yes (Admin) |
| DELETE | `/api/v1/collection/:uid/member/:username/` | Yes (Admin) |
| PATCH | `/api/v1/collection/:uid/member/:username/` | Yes (Admin) |
| POST | `/api/v1/collection/:uid/member/leave/` | Yes |

### Invitations
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/invitation/incoming/` | Yes |
| GET | `/api/v1/invitation/incoming/:uid/` | Yes |
| DELETE | `/api/v1/invitation/incoming/:uid/` | Yes |
| POST | `/api/v1/invitation/incoming/:uid/accept/` | Yes |
| GET | `/api/v1/invitation/outgoing/` | Yes |
| DELETE | `/api/v1/invitation/outgoing/:uid/` | Yes |
| POST | `/api/v1/invitation/outgoing/fetch_user_profile/` | Yes |

### WebSocket
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/api/v1/ws/:ticket/` | Via ticket |

### Health
| Method | Path | Auth Required |
|--------|------|---------------|
| GET | `/health` | No |
| GET | `/ready` | No |
| GET | `/live` | No |

### Debug (DEBUG mode only)
| Method | Path | Auth Required |
|--------|------|---------------|
| POST | `/api/v1/test/authentication/reset/` | No |

---

## Stopping Everything

```bash
# Stop GoatSync
pkill goatsync
# Or press Ctrl+C (graceful shutdown)

# Stop Docker containers
docker compose down

# Stop and remove volumes (wipes data)
docker compose down -v
```

---

## Troubleshooting

### Database Connection Failed
```bash
# Check if PostgreSQL is running
docker compose ps

# Check PostgreSQL logs
docker compose logs postgres

# Test connection manually
psql postgres://goatsync:goatsync@localhost:5432/goatsync
```

### Redis Connection Failed
```bash
# Check if Redis is running
docker compose ps

# Test Redis connection
redis-cli -h localhost -p 6379 ping
```

### Migration Errors
```bash
# Reset database
docker compose down -v
docker compose up -d

# The server will auto-migrate on startup
```

### Port Already in Use
```bash
# Find what's using the port
lsof -i :3735

# Kill it
kill -9 <PID>
```

