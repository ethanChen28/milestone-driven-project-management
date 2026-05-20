# E2E Test Report - Milestone-Driven Project Management

**Date:** 2026-05-19
**Change:** milestone-driven-project-management
**Environment:** Docker Compose (MySQL 8.4, Redis 7.4, Go backend, Vue 3 frontend via Nginx)

---

## 1. Test Execution Results

### Backend Unit/Integration Tests (Go)

| Result | Count |
|--------|-------|
| PASSED | 27 |
| FAILED | 0 |

All 27 backend tests pass, including `TestMVPEndToEnd` which seeds a full workspace and verifies all reporting views.

### Playwright E2E Tests (Chromium)

| Result | Count |
|--------|-------|
| PASSED | 14 |
| FAILED | 0 |
| Duration | 1.8s |

| Test | Result |
|------|--------|
| page loads with correct title and hero section | PASS |
| displays three technology stack cards | PASS |
| shows loading state then portfolio summary | PASS |
| portfolio summary shows numeric values from API | PASS |
| locale switch toggles between Chinese and English | PASS |
| switches back to Chinese after toggling | PASS |
| sets document lang attribute on locale switch | PASS |
| health endpoint returns ok | PASS |
| portfolio dashboard returns valid structure | PASS |
| roadmap overview returns array | PASS |
| ops status returns valid structure | PASS |
| create project and verify it appears in dashboard | PASS |
| create milestone and verify blocked milestone count | PASS |
| generate alerts and verify in weekly review | PASS |

### Frontend Unit Tests (Vitest)

| Result | Count |
|--------|-------|
| PASSED | 2 |
| FAILED | 0 |

---

## 2. New Issues Found

### ISSUE-01 (HIGH): Vitest Scans E2E Directory, Causes Test Runner Conflict

**Location:** `frontend/vite.config.ts` / `frontend/package.json`
**Impact:** `npm run test:unit` (vitest) picks up `e2e/app.spec.ts` and crashes with a Playwright conflict error. CI pipelines that run vitest will fail.

**Evidence:**
```
FAIL  e2e/app.spec.ts
Error: Playwright Test did not expect test.describe() to be called here.
```

**Root Cause:** `vite.config.ts` has no `test.exclude` pattern. Vitest defaults to scanning all `.ts` files, including `e2e/`.

**Fix:** Add to `vite.config.ts`:
```ts
test: {
  exclude: ['e2e/**', 'node_modules/**'],
}
```

---

### ISSUE-02 (HIGH): In-Memory Store Loses All Data on Backend Restart

**Location:** `backend/internal/service/store.go` - `NewStore()` creates empty in-memory maps
**Impact:** All roadmap periods, projects, milestones, work items, alerts, and notifications are lost when the backend container restarts. No data is persisted to MySQL despite MySQL being configured in `docker-compose.yml` and schema migrations existing in `infra/mysql/init/`.

**Evidence:** API responses show data from E2E test runs only (IDs like `prj-013`, `ms-014`), meaning data only survives the current process lifetime.

**Root Cause:** `Store` uses `map[string]T` with `sync.RWMutex`. `Config.MySQLDSN` is loaded but never used. `cmd/server/main.go` creates `NewServer(LoadConfig())` which calls `service.NewStore()` — a pure in-memory constructor.

**Fix:** Implement MySQL-backed `Store` (or at minimum SQLite for MVP) using the existing schema in `001_schema.sql` and `002_gitlab_sync_alerts.sql`.

---

### ISSUE-03 (HIGH): E2E Tests Leave Orphan Data, Dashboard Accuracy Degrades

**Location:** `frontend/e2e/app.spec.ts` - "Data Creation Flow" tests (lines 154-246)
**Impact:** The 3 data-creation tests (`create project`, `create milestone`, `generate alerts`) create permanent records without cleanup. After multiple test runs, dashboard metrics become inflated and unreliable.

**Evidence:** Portfolio dashboard currently shows:
- `activeProjects: 6` (4 are test artifacts)
- `blockedMilestones: 2`, `overdueMilestones: 2` (all from test runs)
- `healthDistribution: {"": 4, "on_track": 2}` — 4 projects with empty health status from test data

**Root Cause:** No `test.afterAll` cleanup, no test database isolation, and in-memory store never resets between test runs.

**Fix Options:**
1. Add `afterAll` to delete created records via API
2. Use a dedicated test workspace or namespace
3. Reset the store between test runs via a `/api/v1/_test/reset` endpoint (test-only)

---

### ISSUE-04 (MEDIUM): Projects Created Without Required Fields

**Location:** `backend/internal/service/store.go` - `CreateProject()`
**Impact:** Projects can be created with empty `status`, `healthStatus`, `owner`, `roadmapItemId`. The database schema marks these as `NOT NULL` but the in-memory store has no validation.

**Evidence:**
```json
{
  "id": "prj-002",
  "name": "Milestone E2E",
  "status": "",
  "healthStatus": "",
  "objective": "",
  "owner": "tester"
}
```

**Root Cause:** `CreateProject()` only checks `PermManageProject` and assigns an ID. No field-level validation like `validateMilestone()` does.

**Fix:** Add validation for required project fields (`name`, `status`, `owner` minimum). Consider adding a `validateProject()` function.

---

### ISSUE-05 (MEDIUM): Notifications Endpoint Has No Authentication

**Location:** `backend/internal/app/server.go:581-602` - `handleNotifications()`
**Impact:** Both `GET /api/v1/notifications` and `POST /api/v1/notifications` require no role header. Any unauthenticated client can list all notifications or send arbitrary notification events.

**Evidence:**
```bash
curl -s -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{"eventType":"spam","target":"victim","title":"Spam","message":"Unwanted"}'
# Returns 200 with created events, no auth required
```

**Fix:** Add RBAC check: `PermViewDashboard` for GET, `PermManageNotification` for POST.

---

### ISSUE-06 (MEDIUM): Sync Failures Endpoint Has No Authentication

**Location:** `backend/internal/app/server.go:484-490` - `handleSyncFailures()`
**Impact:** `GET /api/v1/sync-failures` returns all sync failure records without any role check.

**Fix:** Add `PermManageSyncRule` check (consistent with the resolve endpoint which does have RBAC).

---

### ISSUE-07 (MEDIUM): Alerts Endpoint GET Has No Authentication

**Location:** `backend/internal/app/server.go:546-557` - `handleAlerts()`
**Impact:** `GET /api/v1/alerts` lists all alerts without role verification. While `POST /api/v1/alerts` (generate) has no explicit check either, the `alerts/dismiss` endpoint correctly checks `PermManageAlert`.

**Fix:** Add `PermViewDashboard` for GET, `PermManageAlert` for POST (generate).

---

### ISSUE-08 (MEDIUM): Dashboard Endpoints Don't Verify Read Permissions

**Location:** `backend/internal/app/server.go:303-333`
**Impact:** All dashboard endpoints (`/dashboard/portfolio`, `/dashboard/roadmap`, `/dashboard/project`, `/dashboard/milestone`, `/review/weekly`) can be accessed without any role header. While these are read-only, the spec defines `PermViewDashboard` for viewers.

**Fix:** Add `HasPermission(roleFromHeader(r), PermViewDashboard)` check to dashboard handlers.

---

### ISSUE-09 (LOW): Go Module Declares go 1.13

**Location:** `backend/go.mod`
**Impact:** `go 1.13` is extremely outdated. The code uses generics-capable patterns and the build toolchain may warn or fail on newer Go installations.

**Fix:** Update to `go 1.22` (matches Dockerfile `golang:1.22-alpine`).

---

### ISSUE-10 (LOW): E2E Test Assumes API V1 Prefix but Frontend Doesn't Always Use It

**Location:** `frontend/e2e/app.spec.ts:36` vs `frontend/src/App.vue:23`
**Impact:** E2E tests call `page.request.get("/api/v1/dashboard/portfolio")` directly, bypassing the frontend proxy. This works in CI but doesn't test whether the frontend's `fetch("/api/v1/dashboard/portfolio")` actually works through the proxy chain.

**Evidence:** The frontend's `vite.config.ts` proxies `/api` to `http://127.0.0.1:8080`. In Docker, Nginx proxies `/api` to `backend:8080`. Both work but the E2E test doesn't verify the frontend-to-backend path.

**Fix:** Consider adding at least one test that verifies data flow through the frontend page (the existing `portfolio summary shows numeric values from API` test partially covers this).

---

### ISSUE-11 (LOW): Frontend Has No Error Handling for API Failures

**Location:** `frontend/src/App.vue:21-33`
**Impact:** When backend is unreachable, the `catch` block silently ignores errors and shows `0` for all metrics. No user-visible error state.

**Evidence:**
```ts
} catch {
  // Keep the shell usable even when the backend is not reachable in local static preview.
}
```

**Fix:** Add an error state display (e.g., "Unable to load dashboard data. Retrying...") with a retry button.

---

### ISSUE-12 (LOW): Duplicate Test Data IDs Indicate No Test Isolation

**Location:** `backend/internal/service/store.go:54-57` - `nextID()` uses a global counter
**Impact:** ID collisions or predictable IDs across test runs. The global `sequence` counter in `Store` means IDs are predictable (`prj-001`, `prj-002`, etc.) but shared across all entities, making cross-reference debugging harder.

**Evidence:** Current data shows non-sequential project IDs (`prj-001` through `prj-013`) with gaps, because milestones, work items, etc. also consume from the same counter.

**Fix:** Use UUIDs or prefix-specific counters for better uniqueness guarantees.

---

## 3. Previously Known Issues (from VERIFICATION_REPORT.md)

These issues were already documented and are not new findings:

| # | Issue | Severity | Status |
|---|-------|----------|--------|
| K1 | Webhook no signature verification | MEDIUM | Known, accepted for MVP |
| K2 | Notification adapters are stubs | MEDIUM | Known |
| K3 | No scheduled sync | MEDIUM | Known |
| K4 | Missing data model fields (Description, MergeRequestRefs) | LOW | Known |
| K5 | No auth middleware (X-Role header) | LOW | Known |
| K6 | Hard delete on unlink | LOW | Known |

---

## 4. Summary

### New Issues by Severity

| Severity | Count | Issues |
|----------|-------|--------|
| HIGH | 3 | ISSUE-01 (vitest/e2e conflict), ISSUE-02 (no persistence), ISSUE-03 (test data pollution) |
| MEDIUM | 5 | ISSUE-04 (no project validation), ISSUE-05/06/07/08 (missing auth on 5 endpoints) |
| LOW | 4 | ISSUE-09 (go version), ISSUE-10 (e2e coverage gap), ISSUE-11 (no error UI), ISSUE-12 (ID scheme) |

### Bug Found During Spec Scenario Testing

| Bug | Severity | Description |
|-----|----------|-------------|
| BUG-01 | MEDIUM | `WeeklyReviewView.blockedMilestones` returns `null` instead of `[]` when empty (Go nil slice). See `docs/E2E_TEST_CASES.md` BUG-01. |

### Spec Scenario E2E Results

Full test case execution results: `docs/E2E_TEST_CASES.md`

| Result | Count |
|--------|-------|
| PASS | 20 |
| PASS (caveat) | 2 |
| FAIL | 0 |
| **Total** | **22** |

### Recommendations

**Must fix before MVP:**
1. ISSUE-01: Add vitest exclude pattern (1-line config fix)
2. ISSUE-02: Persist data to MySQL using existing schema
3. ISSUE-03: Add test cleanup or test isolation
4. ISSUE-05/06/07/08: Add RBAC to unauthenticated endpoints

**Should fix before production:**
5. ISSUE-04: Add project field validation
6. All Known issues from VERIFICATION_REPORT.md

### Test Environment Details

```
Services Running:
- mysql:       goal_manager-redis-1   (MySQL 8.4, port 3306)
- redis:       goal_manager-redis-1   (Redis 7.4, port 6379)
- backend:     goal_manager-backend-1 (Go, port 8080)
- frontend:    goal_manager-frontend-1 (Nginx + Vue 3, port 5173)

Playwright:    v1.60.0, Chromium only
Go:            1.22 (Dockerfile)
Node:          20-alpine (Dockerfile)
```
