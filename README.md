# Messenger Svyaz - Go Rewrite

Real-time messenger rewritten from Python/Flask to Go for better performance and deployment simplicity.

## Stack

- **Language:** Go 1.22
- **Router:** chi
- **WebSocket:** gorilla/websocket
- **Database:** PostgreSQL + pgx + sqlc
- **Cache/PubSub:** Redis (go-redis)
- **Auth:** golang-jwt/jwt
- **Config:** godotenv
- **Logging:** slog (std lib)
- **Validation:** go-playground/validator
- **Rate Limiting:** ulule/limiter (Redis backend)

## Project Structure

```
messenger-svyaz/
├── cmd/
│   ├── server/          # HTTP + WebSocket server
│   └── admin/           # CLI: generate-key, list-keys, ban-user, stats
├── internal/
│   ├── config/          # Environment configuration
│   ├── domain/          # Domain entities
│   ├── repository/      # Data access layer (sqlc)
│   ├── storage/         # File storage interface (local/S3)
│   ├── presence/        # Redis-based online status
│   ├── websocket/       # WebSocket hub + Redis Pub/Sub
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   └── db/              # Database connection + migrations
└── docker-compose.yml
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose

### Local Development

1. Clone and install dependencies:
```bash
cd messenger-svyaz-go
go mod tidy
```

2. Start PostgreSQL and Redis:
```bash
docker-compose up postgres redis -d
```

3. Generate sqlc code (after modifying SQL queries):
```bash
cd internal/db
sqlc generate
cd ../..
```

4. Apply migrations:
```bash
docker exec -i messenger-postgres psql -U postgres -d messenger < internal/db/migrations/000001_init_schema.up.sql
```

5. Run server:
```bash
go run cmd/server/main.go
```

6. Check health:
```bash
curl http://localhost:8080/health
```

### Docker Deployment

```bash
docker-compose up -d
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DATABASE_DSN=postgres://postgres:postgres@localhost:5432/messenger?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=change-me-in-production
STORAGE_TYPE=local
STORAGE_PATH=./data
```

## API Endpoints

### Auth
- `POST /login` - User login
- `POST /register` - User registration (requires key)

### Users
- `GET /api/users/list` - List users
- `GET /api/profile` - Current user profile
- `POST /api/profile/update` - Update profile
- `POST /api/profile/change_password` - Change password
- `GET /api/avatar/<username>` - Get avatar
- `POST /api/upload/avatar` - Upload avatar

### Chat (P2P)
- `GET /api/chat/messages/<other_user>` - Chat history
- `GET /api/chat/messages/saved` - Saved messages
- `POST /api/chat/send` - Send message
- `POST /api/chat/save` - Save message
- `GET /api/chat/recent` - Recent chats

### Groups
- `GET /api/groups/my` - My groups
- `POST /api/groups/create` - Create group
- `POST /api/groups/join` - Join group
- `GET /api/group/messages/<group_id>` - Group messages
- `GET /api/group/info/<group_id>` - Group info
- `POST /api/group/send` - Send to group
- `POST /api/group/reaction` - Add reaction

### Support
- `GET /api/support/messages` - Support messages
- `POST /api/support/send` - Send to support

## WebSocket Events

- `connect/disconnect` - Connection events
- `join_chat` - Join P2P chat room
- `join_group` - Join group room
- `typing` - Typing indicator
- `new_message` - New P2P message
- `new_group_message` - New group message

## CLI Commands

```bash
# Generate registration keys
go run cmd/admin/main.go generate -n 5

# List all keys
go run cmd/admin/main.go list-keys

# Ban user
go run cmd/admin/main.go ban-user <username>

# Show stats
go run cmd/admin/main.go stats
```

## Development Status

- [x] Stage 1: Analysis
- [x] Stage 2: Skeleton (config, DB, health-check)
- [ ] Stage 3: Domain & Database (migrations, sqlc, repositories)
- [ ] Stage 4: HTTP API (auth, users, chats, groups)
- [ ] Stage 5: WebSocket Hub (Redis Pub/Sub, multi-instance)
- [ ] Stage 6: Final assembly (Docker, README)

## Notes

- **ID Generation:** 8-character base32 ULID (32^8 = 1.1T combinations)
- **Online Status:** Redis-based with TTL (not in DB)
- **Rate Limiting:** HTTP via ulule/limiter (Redis), WebSocket via token bucket
- **Storage:** Local filesystem with interface for future S3 migration
- **No P2P:** P2P functionality deferred to v2
- **API Compatible:** Maintains compatibility with C# client
