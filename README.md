# Spotify Insights

A personal, always-on "Spotify Wrapped." Continuously collects your listening
history (Spotify only exposes a rolling window), then turns it into a live
dashboard, a swipeable shareable story, and an embeddable "now playing" widget.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full system design,
data model, API contract, and UI/UX spec — read that before opening a PR, it's
the source of truth for how pieces fit together.

**Status:** Phase 1 (backend core) and Phase 2 (frontend) are both built —
OAuth/sessions, poller, stats API, widget endpoint, the React dashboard/story
UI, and a PWA manifest + service worker. Capacitor iOS/Android projects are
scaffolded (`apps/web/android`, `apps/web/ios`) but unbuilt/unverified — that
needs Xcode and the Android SDK, which this environment doesn't have.

## Stack

- **Backend:** Go (stdlib `net/http`, Go 1.22+ enhanced `ServeMux` — no router
  dependency), Postgres
- **Frontend:** React (Vite, TypeScript), wrapped with Capacitor for iOS/Android
- **Delivery:** installable PWA + native mobile shell

## Getting started

Prerequisites: Docker + Docker Compose, Node 20+, a Spotify Developer app
(Development Mode is fine — see the note on Spotify API limits below).

```bash
# 1. Configure
cp .env.example .env
# fill in SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, SPOTIFY_REDIRECT_URI,
# SESSION_SECRET, TOKEN_ENCRYPTION_KEY, WIDGET_HMAC_SECRET
#
# Spotify requires the loopback IP literal (127.0.0.1), not the "localhost"
# hostname, for redirect URIs — register http://127.0.0.1:8080/auth/callback
# on the app's dashboard settings to match SPOTIFY_REDIRECT_URI exactly.

# 2. Bring up Postgres + the Go API + the worker
docker compose up --build

# 3. Start the frontend (separate terminal)
cd apps/web && npm install && npm run dev
```

Open `http://127.0.0.1:5173` (not `localhost` — the OAuth callback lands on
`127.0.0.1`, and the login session cookie won't carry over if the two
hostnames don't match), click **Connect Spotify**. The Vite dev server
proxies `/api`, `/auth`, and `/widget` to the Go API on `:8080` (see
`apps/web/vite.config.ts`) — this keeps the session cookie same-origin in dev,
matching how Caddy routes things in production.

Migrations run automatically on `docker compose up` (via `api`'s `-migrate`
flag). Rebuild after Go code changes with `docker compose up --build`.

### Mobile build

```bash
cd apps/web
npm run build
npx cap sync
npx cap open ios      # or: npx cap open android
```

Requires Xcode (iOS) or Android Studio (Android) — not available in every dev
environment. The `ios/` and `android/` projects are already generated; this
step just opens them in their native IDE to build/run on a simulator/emulator
or a device.

### Production deploy

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
```

Set `FRONTEND_URL` in `.env` to your real domain first — Caddy serves the API
and the built web app from one origin in production, so `/auth/callback`
should redirect there instead of to the dev-only Vite address.

## Project structure

```
apps/api/     Go backend — cmd/api (HTTP server), cmd/worker (poller + rollup)
apps/web/     React frontend (Vite + TypeScript), PWA, wrapped for mobile via Capacitor
docs/         Architecture and design documentation
```

## A note on Spotify API limits

New Spotify Developer apps run in Development Mode: 5 authorized users per app,
owner must have Spotify Premium, and several endpoints (audio features, audio
analysis, recommendations, related artists) are unavailable to apps created
after November 2024. This project is built around what's actually available —
see `docs/ARCHITECTURE.md` §2 for specifics before adding a feature that
assumes a deprecated endpoint.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md).
