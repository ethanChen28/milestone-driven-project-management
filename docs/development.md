# Development Notes

## Backend Storage

The backend supports two storage modes:

- `STORAGE_BACKEND=memory`: in-memory repository for unit tests and quick local development. Data is lost when the backend exits.
- `STORAGE_BACKEND=mysql`: MySQL-backed repository for Docker/runtime durability. `MYSQL_DSN` is required and startup fails if MySQL cannot be reached.

Default local `LoadConfig()` behavior is `memory` so backend unit tests do not require MySQL. Docker Compose sets `STORAGE_BACKEND=mysql` and uses:

```bash
MYSQL_DSN=goal:goal@tcp(mysql:3306)/goal_manager?parseTime=true
REDIS_ADDR=redis:6379
APP_DEFAULT_LOCALE=zh-CN
```

Run the full Docker stack with:

```bash
docker compose up --build
```

## MySQL Integration Tests

Backend MySQL persistence tests are guarded by `MYSQL_INTEGRATION_DSN`:

```bash
cd backend
MYSQL_INTEGRATION_DSN='goal:goal@tcp(127.0.0.1:3306)/goal_manager?parseTime=true' go test ./internal/app -run MySQLPersistence
```

If `MYSQL_INTEGRATION_DSN` is not set, the integration test is skipped and regular unit tests continue to use memory storage.

## E2E With Alternate Backend Port

When port `8080` is unavailable, run the backend on another port and point Vite's proxy at it:

```bash
cd backend
STORAGE_BACKEND=memory APP_LISTEN_ADDR=:18080 go run ./cmd/server

cd ../frontend
API_PROXY_TARGET=http://127.0.0.1:18080 npm run dev -- --host 127.0.0.1 --port 15173
E2E_BASE_URL=http://127.0.0.1:15173 npm run test:e2e
```
