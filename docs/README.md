# Spotify Insights

A personal, always-on "Spotify Wrapped." Continuously collects your listening
history (Spotify only exposes a rolling window), then turns it into a live
dashboard, a swipeable shareable story, and an embeddable "now playing" widget.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full system design,
data model, API contract, and UI/UX spec — read that before opening a PR, it's
the source of truth for how pieces fit together.

## Stack

- **Backend:** Go (Gin), Postgres
- **Frontend:** React (Vite), wrapped with Capacitor for iOS/Android
- **Delivery:** installable PWA + native mobile shell

## Getting started

Prerequisites: Docker + Docker Compose, Node 20+ (for the frontend, which runs
outside the compose stack for fast reloads), a Spotify Developer app (Development
Mode is fine — see the note on Spotify API limits below).

```bash
# 1. Configure
cp .env.example .env
# fill in SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, SPOTIFY_REDIRECT_URI
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
hostnames don't match), click **Connect Spotify**.

Migrations run automatically on `docker compose up`. Rebuild after Go code changes
with `docker compose up --build`.

### Mobile build

```bash
cd apps/web
npm run build
npx cap sync
npx cap open ios      # or: npx cap open android
```

## Project structure

```
apps/api/     Go backend (API server + worker binary)
apps/web/     React frontend (web, PWA, wrapped for mobile via Capacitor)
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

## License

MIT — see `LICENSE`.
