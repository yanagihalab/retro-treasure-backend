# RELIC RAID

RELIC RAID is a Go-based web game service with an embedded browser game UI. It serves the API and static game screens from one binary, and can run locally, inside Docker, in an Android WebView, or behind nginx on a VPS.

The current game is a landscape-oriented, Cthulhu-inspired collection RPG about disaster-prevention relic cards, checkpoint nodes, and endurance-style boss battles.

## Current Features

- Embedded web UI under `/static/`
- Optional base path support such as `/games/`
- REST API for authentication, player state, cards, bosses, gacha, checkpoints, inventory, and notices
- 100 player cards across `heart`, `tech`, and `body` attributes
- Boss selection and endurance battle flow
- Boss attack attributes, attack skills, hints, and battle status effects
- Checkpoint map page with QR node progression
- Android WebView project under `android/`
- File-backed persistence for users, cards, decks, progress, tickets, checkpoint history, and tokens
- bcrypt password hashing, with legacy SHA-256 login compatibility
- VPS deployment assets under `deploy/`

## Repository Layout

```text
.
├─ cmd/server/                 # HTTP server entrypoint
├─ internal/config/            # Environment configuration
├─ internal/handler/           # HTTP handlers
├─ internal/middleware/        # Auth middleware
├─ internal/model/             # API and game models
├─ internal/repository/        # In-memory repository with JSON persistence
├─ internal/seed/              # Master data seeds
├─ internal/service/           # Game and API business logic
├─ internal/webassets/         # Embedded HTML/CSS/JS/images
├─ android/                    # Android WebView wrapper
├─ deploy/                     # VPS systemd/nginx deployment files
├─ migrations/                 # Future SQL migration draft
├─ Dockerfile
├─ docker-compose.yml
└─ README.md
```

## Requirements

- Go 1.23 or newer for local development
- Node.js for JavaScript syntax checks
- Docker and Docker Compose, optional
- Android Studio, optional

The VPS deployment uses a prebuilt Linux binary, so Go does not need to be installed on the VPS.

## Quick Start

Run the service locally:

```bash
go run ./cmd/server
```

Open:

- Game home: `http://localhost:8080/static/`
- Health check: `http://localhost:8080/health`

Expected health response:

```json
{"status":"ok"}
```

## Environment Variables

| Name | Default | Description |
| --- | --- | --- |
| `APP_NAME` | `retro-treasure-api` | Name shown in server logs |
| `APP_HOST` | empty | Bind host. Use `127.0.0.1` behind nginx |
| `APP_PORT` | `8080` | Internal HTTP port |
| `APP_BASE_PATH` | empty | Public path prefix, for example `/games` |
| `DATA_DIR` | empty | Directory for persistent state |
| `APP_STATE_FILE` | empty | Explicit JSON state file path |

If `APP_STATE_FILE` is empty and `DATA_DIR` is set, the server writes state to:

```text
${DATA_DIR}/state.json
```

Local persistent run:

```bash
mkdir -p .local-data
DATA_DIR=.local-data go run ./cmd/server
```

Base path run matching the VPS:

```bash
mkdir -p .local-data
APP_HOST=127.0.0.1 \
APP_PORT=8080 \
APP_BASE_PATH=/games \
DATA_DIR=.local-data \
go run ./cmd/server
```

Then open `http://localhost:8080/games/`.

## Base Path Behavior

The app can be published under a sub-path such as `/games/`.

When `APP_BASE_PATH=/games` is set:

- `/games/` serves the game home.
- `/games/static/...` serves static assets.
- `/games/api/...` routes to the same API handlers as `/api/...`.
- Served HTML/JS/CSS receives `window.__APP_BASE_PATH__` and a matching `<meta name="app-base-path">`.

This removes the need to rely on nginx `sub_filter` for asset and API path rewriting.

## API Examples

Register:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"player1","password":"password123"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"player1","password":"password123"}'
```

Use the returned token:

```bash
curl http://localhost:8080/api/player/me \
  -H 'Authorization: Bearer <token>'
```

With `APP_BASE_PATH=/games`, use `/games/api/...`:

```bash
curl http://localhost:8080/games/api/checkpoints/master
```

## Main API Routes

### Auth

- `POST /api/auth/register`
- `POST /api/auth/login`

### Player

- `GET /api/player/me`

### Exploration and Items

- `GET /api/areas`
- `POST /api/explore`
- `GET /api/items/inventory`
- `GET /api/encyclopedia`

### Cards and Gacha

- `GET /api/cards/me`
- `GET /api/cards/deck`
- `GET /api/cards/collection`
- `GET /api/cards/archive`
- `POST /api/cards/upgrade`
- `POST /api/cards/deck`
- `POST /api/gacha/draw`

### Bosses

- `GET /api/boss`
- `POST /api/boss/auto`

### Checkpoints

- `GET /api/checkpoints/master`
- `GET /api/checkpoints/history`
- `POST /api/checkpoints/claim`

### Daily and Notices

- `POST /api/login-bonus/claim`
- `GET /api/notices`
- `GET /health`

## Persistence

Runtime data is stored in memory and periodically serialized to a JSON state file after mutations. This keeps the implementation simple while preventing VPS restarts from wiping players and progress.

Persisted state includes:

- Users and password hashes
- Auth tokens
- Player status
- Inventory and encyclopedia progress
- Owned cards, deck slots, upgrades, and tickets
- Exploration logs
- Login bonus claim state
- Checkpoint claim history

The state file should live outside the repository, for example:

```text
/home/ubuntu/.relic-raid/data/state.json
```

Back it up before replacing production binaries or changing persistence code.

## Authentication

New passwords are hashed with bcrypt.

Older SHA-256 hashes are still accepted for login compatibility. When a legacy hash login succeeds, the password hash is upgraded to bcrypt and written back to the persistent state file.

Never commit production state files or environment files.

## Docker

Build:

```bash
docker build -t relic-raid .
```

Run without persistence:

```bash
docker run --rm -p 8080:8080 relic-raid
```

Run with persistence:

```bash
mkdir -p .local-data
docker run --rm \
  -p 8080:8080 \
  -e DATA_DIR=/data \
  -v "$PWD/.local-data:/data" \
  relic-raid
```

Docker Compose:

```bash
docker compose up --build
```

Stop:

```bash
docker compose down
```

## Android

Open `android/` in Android Studio.

For emulator development, start the Go server locally:

```bash
go run ./cmd/server
```

The emulator uses:

```text
http://10.0.2.2:8080/static/
```

For physical devices, copy the local properties example and set a reachable LAN or HTTPS URL:

```bash
cp android/local.properties.example android/local.properties
```

```properties
GAME_BASE_URL=http://<your-mac-lan-ip>:8080/static/
```

For production WebView builds, point `GAME_BASE_URL` to:

```text
https://ik1-206-76937.vs.sakura.ne.jp/games/
```

See `android/README.md` for Android-specific notes.

## VPS Deployment

The Sakura VPS deployment publishes RELIC RAID under:

```text
https://ik1-206-76937.vs.sakura.ne.jp/games/
```

Runtime design:

- nginx exposes only ports 80 and 443.
- RELIC RAID binds to `127.0.0.1:8090`.
- Existing Tozan Todoke ports `8080` and `8788` are not reused.
- Environment and state live under `/home/ubuntu/.relic-raid/`.
- systemd service name is `relic-raid.service`.

Deployment files are tracked in:

```text
deploy/relic-raid.service
deploy/nginx-games-location.conf
deploy/README.md
```

See `deploy/README.md` for the exact install, update, nginx, systemd, and verification commands.

## Production Build

Build a Linux binary for the VPS:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o relic-raid ./cmd/server
```

Deploy the binary and tracked deploy files:

```bash
rsync -av relic-raid deploy/relic-raid.service deploy/nginx-games-location.conf \
  sakura-tozantodoke:/home/ubuntu/relic-raid/
```

Then apply the sudo steps from `deploy/README.md` on the VPS.

## Verification

Run Go tests:

```bash
go test ./...
```

Check JavaScript syntax:

```bash
for f in internal/webassets/static/js/*.js; do
  node --check "$f"
done
```

Local base path smoke test:

```bash
rm -rf /tmp/relic-raid-verify
mkdir -p /tmp/relic-raid-verify
APP_HOST=127.0.0.1 \
APP_PORT=18090 \
APP_BASE_PATH=/games \
DATA_DIR=/tmp/relic-raid-verify \
go run ./cmd/server
```

In another terminal:

```bash
curl -I http://127.0.0.1:18090/games/
curl -I http://127.0.0.1:18090/games/static/js/app.js
curl http://127.0.0.1:18090/games/api/checkpoints/master
```

## Development Notes

- Keep secrets, production env files, and state files outside git.
- Prefer adding game master data in `internal/seed/seed.go`.
- Keep API response shapes in `internal/model/`.
- Keep browser UI assets in `internal/webassets/static/`.
- The current JSON persistence is suitable for a small single-process VPS deployment. For multi-instance or larger production use, move persistent state to SQLite or PostgreSQL.
