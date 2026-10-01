# Caddio — Self-Hosted Caddy Server Management

**Caddio** is a production-quality self-hosted web interface for managing, monitoring, and configuring a Caddy web server safely.

---

## 1. Key Features

### 🛡️ Phase 1: Foundation, Auth & Infrastructure
- **First-Run Admin Setup**: Secure setup wizard for initial administrator creation.
- **SQLite Session Authentication**: Opaque 32-byte cryptographically secure tokens stored in SQLite with WAL mode and delivered via HTTP-only cookies (`caddio_session`).
- **Host Caddy Binary Version Detection**: Direct binary version discovery (`caddy version`) via `os/exec` with context timeouts (`GET /api/system/caddy/version`).
- **Operational Dashboard & Audit Log**: Aggregated system metrics, Caddy connectivity status, site summaries, and audit trail stream (`GET /api/audit-logs`).

### 🌐 Phase 2: Reverse Proxy & Site Management
- **Automatic Route Discovery**: Scans all Caddy HTTP server blocks (`apps.http.servers`) and Caddyfile-generated subroutes (`handler: "subroute"`).
- **Caddio Route Ownership Scoping**: Assigns deterministic route IDs (`caddio_site_<slug>_<hash>`) and ownership markers (`managed_by: caddio`). External/unmanaged Caddy routes are protected in read-only mode (`403 Forbidden` on unauthorized edits).
- **Pre-flight Dry-run Validation**: Dry-run route validation endpoint (`POST /api/sites/validate`) checking domains, schemes, and upstreams before applying.
- **Unrelated Config Preservation**: Guarantees non-managed server blocks, custom TLS automation rules, file servers, and custom routes remain untouched during site operations.

### ⚙️ Phase 3: Configuration Workspace, Diff & Safe Apply
- **CodeMirror 6 JSON Editor**: Feature-rich editor (`/config`) with JSON syntax verification, auto-formatting, line numbers, and copy tools.
- **Backend Structured Line Diff Engine**: Computes line-by-line unified diffs with exact addition (+), deletion (-), and modification (~) statistics before applying.
- **Optimistic Concurrency Protection**: Compares client `base_hash` against active Caddy `current_hash` on server, returning `409 CONFIG_CONFLICT` if configuration changed elsewhere.
- **SQLite Snapshot Store & History**: Automatic pre-apply backup snapshots in SQLite (`config_snapshots`). Dedicated history page (`/config/history`) with snapshot inspection and instant rollback dialogs.
- **Post-Apply Hash Verification**: Re-fetches Caddy configuration after `POST /load` to verify active hash matches requested hash (returning `500 APPLY_VERIFICATION_FAILED` on mismatch).
- **Dual Light/Dark Theme**: Dark-first visual design with full support for Light Theme across all cards, toolbars, line diffs, tables, and dialogs.

---

## 2. Architecture Overview

```text
Internet
   │
   ▼
 Caddy Server (Port 80/443)
   │
   ├─────── internal Admin API (http://caddy:2019) ─────────┐
   │                                                         ▼
 Host Caddy Binary (/usr/bin/caddy) ─── os/exec ──► Caddio Manager (Single Binary)
                                                             │
                                                             ▼
                                                    SQLite Database (/data/caddy-manager.db)
```

- **Single Embedded Binary**: Svelte 5 SPA assets are compiled into `backend/frontend_dist` and embedded into the Go binary using `go:embed`.
- **Host Binary Exec**: Direct execution of `/usr/bin/caddy` for version discovery.
- **Isolated Admin API**: Communicates with Caddy's Admin API (`http://127.0.0.1:2019`) over internal networks; Admin API port 2019 is never exposed publicly.

---

## 3. Technology Stack

### Backend
- **Go 1.24+**: High-performance backend engine.
- **Fiber v2**: Fast HTTP web framework.
- **SQLite (`modernc.org/sqlite`)**: Pure Go CGO-free SQLite driver with WAL mode and foreign key constraints enabled.
- **Bcrypt**: Password hashing.

### Frontend
- **Svelte 5**: Reactive UI components using Svelte 5 runes (`$state`, `$derived`, `$effect`).
- **CodeMirror 6**: Code editor with JSON syntax highlighting (`@codemirror/lang-json`, `@codemirror/theme-one-dark`).
- **TypeScript**: Static typing across models and API contracts.
- **Vite 8**: Fast production building.
- **Tailwind CSS v4**: Styling and dual light/dark theme support.
- **Lucide Icons**: Vector iconography.

---

## 4. Environment Variables

| Variable | Default | Required in Prod | Description |
|---|---|---|---|
| `APP_ENV` | `development` | No | `development` or `production` |
| `PORT` | `8080` | No | HTTP listening port (1-65535) |
| `DATABASE_PATH` | `./caddy-manager.db` | No | SQLite database file path |
| `CADDY_ADMIN_URL` | `http://localhost:2019` | No | Internal Caddy Admin API endpoint |
| `CADDY_BINARY_PATH` | `/usr/bin/caddy` | No | Path to host Caddy binary executable |
| `SESSION_SECRET` | auto-generated (dev) | **Yes** (in prod) | Secret key for session cookie signing (min 32 chars in prod) |

---

## 5. API Endpoints

### Public & Infrastructure
- `GET /health` : Liveness probe.
- `GET /ready` : Readiness probe (SQLite accessibility & Caddy status).
- `GET /api/system/caddy/version` : Returns host Caddy binary version.

### Authentication
- `POST /api/auth/setup` : First-run admin registration.
- `POST /api/auth/login` : User authentication & session cookie generation.
- `POST /api/auth/logout` : Session termination.
- `GET  /api/auth/me` : Current authentication state.

### Operational Dashboard & Audit
- `GET /api/dashboard` : Aggregated metrics, site stats, and recent audit activity.
- `GET /api/audit-logs` : Retrieve system audit log history.

### Site Management (`/api/sites`)
- `GET    /api/sites` : Discover all routes (managed sites & unmanaged route stats).
- `GET    /api/sites/:id` : Get single site details.
- `POST   /api/sites/validate` : Dry-run pre-flight validation.
- `POST   /api/sites` : Create new reverse proxy site and apply to Caddy.
- `PUT    /api/sites/:id` : Update existing Caddio-managed site (`403` on unmanaged routes).
- `DELETE /api/sites/:id` : Remove Caddio-managed route (`403` on unmanaged routes).

### Configuration Workspace (`/api/config`)
- `GET    /api/config` : Fetch active Caddy JSON, active SHA-256 hash, and host version.
- `POST   /api/config/validate` : Dry-run syntax & Caddy structure validation with optimistic conflict check.
- `POST   /api/config/diff` : Compute backend line diff and change stats comparing active vs proposed JSON.
- `POST   /api/config/apply` : Validate, snapshot to SQLite, check optimistic concurrency (`base_hash`), execute `POST /load`, and verify post-apply hash.
- `GET    /api/config/history` : List historical configuration snapshot metadata.
- `GET    /api/config/history/:id` : Fetch full details and raw JSON of a specific snapshot.
- `POST   /api/config/history/:id/rollback` : Restore historical configuration snapshot to Caddy.

---

## 6. Local Development & Building

### Prerequisites
- **Go 1.24+**
- **Node.js 20+** / **Bun 1.2+**

### Local Development Steps

1. **Build Frontend**:
   ```bash
   cd frontend
   bun install
   bun run build
   ```

2. **Sync Frontend Dist & Build Backend**:
   ```bash
   cd ..
   mkdir -p backend/frontend_dist
   cp -r frontend/dist/* backend/frontend_dist/
   cd backend
   go build -o bin/caddio ./cmd/server
   ```

3. **Run Tests**:
   ```bash
   cd backend
   go test -v ./...
   ```

4. **Run Application**:
   ```bash
   ./bin/caddio
   ```

---

## 7. Docker & Docker Compose Deployment

### Docker Compose

```yaml
services:
  caddio:
    image: registry.devteddy.my.id/caddio:latest
    container_name: caddio
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_PATH=/data/caddy-manager.db
      - CADDY_ADMIN_URL=http://caddy:2019
      - CADDY_BINARY_PATH=/usr/bin/caddy
      - APP_ENV=production
      - SESSION_SECRET=change_this_to_a_secure_random_32_character_string_in_production
      - TZ=Asia/Jakarta
    volumes:
      - manager_data:/data
      - /usr/bin/caddy:/usr/bin/caddy:ro

volumes:
  manager_data:
```

### Docker Build

```bash
docker build -t caddio:latest .
```

---

## 8. License

MIT License.
