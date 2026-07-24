# Spotify Insights — Architecture & Design Spec

**Purpose of this document:** a single source of truth for Claude Code (or any dev) to
build this project against. Covers system architecture, file organization, data model,
API contract, and UI/UX design direction. No code here — this is the blueprint.

---

## 1. Product summary

A personal, always-on "Spotify Wrapped" that:
1. Continuously collects listening history Spotify itself doesn't retain long-term.
2. Renders it as a **dashboard** (trends over time).
3. Renders it as a **story** (swipeable, shareable, Wrapped-style slides).
4. Exposes a **public embeddable widget** (live "now playing" badge for a README/site).

Stack: **Go (Gin)** backend, **React** frontend, **Postgres** storage.
Delivery: **responsive PWA + native mobile shell via Capacitor** — one React
codebase covers desktop/tablet/mobile web *and* gets packaged as a real iOS/Android
app, without maintaining a second frontend.

---

## 2. System architecture

```
                        ┌─────────────────────────┐
                        │        Spotify API       │
                        └────────────┬─────────────┘
                                     │ OAuth (PKCE) + polling
                                     ▼
┌────────────────────────────────────────────────────────────────┐
│                         Go backend (single binary,              │
│                         two run-modes: api / worker)            │
│                                                                  │
│  cmd/api ─────────► HTTP server (Gin)                           │
│    ├─ /auth/*        OAuth login, callback, refresh              │
│    ├─ /api/stats/*   dashboard data (auth required)              │
│    ├─ /api/story     story-slide data (auth required)            │
│    └─ /widget/:id    public, token-signed, cached                │
│                                                                  │
│  cmd/worker ──────► Scheduled poller (ticker, e.g. every 20 min) │
│    └─ pulls recently-played + top-tracks per user, writes to DB  │
└───────────────────────────────┬──────────────────────────────────┘
                                │
                                ▼
                        ┌───────────────┐
                        │   Postgres     │
                        └───────────────┘
                                ▲
                                │ REST (JSON)
┌───────────────────────────────┴──────────────────────────────────┐
│                         React app (Vite + PWA)                    │
│   /login   /dashboard   /story   (widget is server-rendered,      │
│                                    not part of the React app)      │
│                                                                     │
│   Same build wrapped by Capacitor ──► iOS app / Android app        │
└────────────────────────────────────────────────────────────────────┘
```

**Mobile:** `apps/web` is wrapped, not rewritten. Capacitor gives it a native
WebView shell plus access to native APIs if you ever want them (share sheet for
story export, push notifications, etc.) — but day one it's literally the same
dashboard/story/login code, built once, deployed three ways (web, iOS, Android).

Why one Go binary with two modes instead of two services: simpler deploy for a
portfolio project, still lets you *talk about* the separation in interviews. Split
into a real second deployable later only if you want to demonstrate that explicitly.

**Reverse proxy:** Caddy in front of both the Go binary and the built React static
files — one TLS cert, one domain, `/api/*` and `/widget/*` routed to Go, everything
else served as static files.

---

## 3. File organization (monorepo)

```
spotify-insights/
├── apps/
│   ├── api/
│   │   ├── cmd/
│   │   │   ├── api/main.go            # HTTP server entrypoint
│   │   │   └── worker/main.go         # Poller entrypoint
│   │   ├── internal/
│   │   │   ├── auth/                  # OAuth + PKCE, JWT session issuing
│   │   │   ├── spotify/               # Spotify API client wrapper
│   │   │   ├── stats/                 # Aggregation queries → dashboard/story DTOs
│   │   │   ├── widget/                # Widget rendering (SVG/HTML) + caching
│   │   │   ├── poller/                # Scheduled collection logic
│   │   │   ├── db/                    # sqlc/pgx queries, migrations runner
│   │   │   └── config/                # env config loading
│   │   ├── migrations/                # SQL migration files
│   │   └── go.mod
│   │
│   └── web/
│       ├── src/
│       │   ├── pages/                 # Login, Dashboard, Story
│       │   ├── components/
│       │   │   ├── charts/            # Chart.js wrapper components
│       │   │   ├── story/             # Slide components + swipe engine
│       │   │   └── ui/                # Buttons, cards, layout primitives
│       │   ├── lib/                   # API client, auth context
│       │   ├── styles/                # Design tokens (see §5), global CSS
│       │   └── manifest.json          # PWA manifest
│       ├── ios/                       # Capacitor-generated native project
│       ├── android/                   # Capacitor-generated native project
│       ├── capacitor.config.ts
│       └── package.json
│
├── docker-compose.yml                 # postgres + api + worker, one command up
├── Caddyfile                          # reverse proxy: /api, /widget → Go; else → static web
├── docs/
│   └── ARCHITECTURE.md                # this file
├── README.md                          # onboarding for open-source contributors
├── CONTRIBUTING.md                    # contribution + commit conventions
└── .env.example
```

Local dev: `docker compose up` brings up Postgres + the Go API + the worker together;
`npm run dev` in `apps/web` runs the frontend separately against them (fastest
iteration loop). `docker compose up --build` picks up Go code changes. Caddy sits
in front only in the production compose profile — not needed for local dev.

---

## 4. Data model (Postgres)

```
users
  id            uuid pk
  spotify_id    text unique
  display_name  text
  created_at    timestamptz

spotify_tokens
  user_id        uuid fk -> users
  access_token    text (encrypted at rest)
  refresh_token   text (encrypted at rest)
  expires_at      timestamptz

listening_events              -- raw poll results, append-only
  id            bigserial pk
  user_id       uuid fk
  track_id      text
  track_name    text
  artist_id     text
  artist_name   text
  played_at     timestamptz
  source        text          -- 'recently_played' | 'currently_playing'

artist_genres                 -- cached lookups, refreshed periodically
  artist_id     text pk
  genres        text[]

top_snapshots                 -- daily rollups for fast dashboard queries
  user_id       uuid fk
  period        text          -- 'short_term' | 'medium_term' | 'long_term'
  snapshot_date date
  top_artists   jsonb
  top_tracks    jsonb

widget_tokens
  token         text pk       -- HMAC-signed, embeddable in a public URL
  user_id       uuid fk
  created_at    timestamptz
  revoked       boolean
```

Rollup jobs (in `worker`) aggregate `listening_events` into `top_snapshots` nightly
so dashboard/story reads never scan raw event rows.

---

## 5. API contract (summary)

| Method | Path                     | Auth   | Purpose |
|---|---|---|---|
| GET  | `/auth/login`            | none   | Redirect to Spotify OAuth |
| GET  | `/auth/callback`         | none   | Exchange code, create session, set cookie |
| POST | `/auth/refresh`          | session| Refresh access token |
| GET  | `/api/stats/top-artists` | session| Top artists, by time range |
| GET  | `/api/stats/genres`      | session| Genre distribution + drift |
| GET  | `/api/stats/heatmap`     | session| Listening by day/hour |
| GET  | `/api/story`             | session| Ordered slide payloads for story view |
| POST | `/api/widget/token`      | session| Issue/rotate a widget token |
| GET  | `/widget/:token`         | none (signed) | Returns cached SVG/HTML "now playing" |

Widget endpoint: verify HMAC signature on `:token`, check a short-TTL cache (in-memory
`sync.Map` or Redis if you want that on the resume) before calling Spotify, to avoid
rate-limit exposure from a public, cacheable, unauthenticated endpoint.

---

## 6. UI/UX design direction

Design brief, treated the way a founder obsessed with conversion and a designer who's
studied Linear/Superhuman/Vercel/Raycast/Arc would treat it: this product has exactly
one moment that matters — the click that connects a Spotify account — and everything
before that click exists to earn it. Everything after it exists to make the person
feel like they're seeing something about themselves nobody else could show them.

### Design tokens

**Color** — avoid the three defaults every AI-generated page reaches for (cream +
serif + terracotta; near-black + acid-green; broadsheet hairlines). This product's
subject is *sound as data* — frequency, rhythm, signal — so the palette comes from
an analog mixing desk, not from "music app" cliché (no Spotify green, no neon).

| Token | Hex | Use |
|---|---|---|
| `bg-base` | `#141217` | App background — warm-black, not pure black |
| `bg-raised` | `#1D1A21` | Cards, panels |
| `signal-coral` | `#FF6B4A` | Primary accent — CTAs, active states, the waveform |
| `signal-cyan` | `#5FE3C4` | Secondary accent — secondary data series, links |
| `ink-100` | `#F2EFEC` | Primary text |
| `ink-400` | `#8B8790` | Secondary text, captions |

**Type**
- Display: **Space Grotesk** (geometric, slightly mechanical — reads like signage on
  studio gear, not a lifestyle brand). Used at large sizes only, bold weight.
- Body: **Inter** — gets out of the way.
- Data/mono: **JetBrains Mono** — for numbers, timestamps, BPM-style stats. Numbers in
  this product are the content, not decoration, so they get their own typeface.

**Layout concept** — "the mixing console." The dashboard is built from modular
channel-strip cards, not a generic grid of stat tiles:

```
┌──────────────┬──────────────┬──────────────┐
│  TOP ARTIST   │   GENRES      │  TIME-OF-DAY  │
│  ▁▂▅▇▅▂▁      │  ◐ donut      │  ▓▓░░▓▓░░     │
│  channel 01   │  channel 02   │  channel 03   │
└──────────────┴──────────────┴──────────────┘
```
Each card is a "channel" with a small monospace label (`CH.01`) — this is one of the
few cases where numbered markers are earned, because the console metaphor makes the
numbering mean something (channel order), not decorative sequencing.

**Signature element** — a live animated waveform that:
- Plays as the hero visual on the landing/login screen (idle animation, subtle)
- Becomes the loading state everywhere data is fetched (never a generic spinner)
- Appears as the divider between dashboard sections
- Pulses in `signal-coral` when data is "hot" (e.g. currently playing on the widget)

This is the one place the design spends its boldness. Everything else — spacing,
card treatment, chart styling — stays quiet and disciplined around it.

### Landing / login screen (the only screen a stranger sees)

Single job: get the Spotify connect click. No feature list, no pricing (there's
nothing to sell), no scrolling required to find the CTA.

- Hero: the waveform signature, live, ambient — under it, one line stating exactly
  what happens next in plain terms: "Connect Spotify. We start listening to your
  listening." Not "unlock insights" — say what happens.
- One button: **Connect Spotify** in `signal-coral`. No secondary CTA competing with it.
- Below the fold (optional, not required to act): three small proof slides — a
  blurred/sample story card, a sample widget embed, a sample dashboard chart — so a
  skeptical visitor can see the payoff before clicking, without needing to.

### Dashboard

- Mixing-console grid of channel cards (§ layout above), responsive: 3 columns
  desktop → 2 columns tablet → 1 column mobile, cards keep their internal proportions
  rather than stretching.
- Time-range switch (short/medium/long term) styled as a physical-feeling toggle,
  not a dropdown — this is a console, controls should feel like controls.

### Story view

- Full-bleed, one slide per screen, swipe (touch) / arrow-key (desktop) navigation.
- Each slide: one number or one chart, large, plus one line of plain-language
  context — never more than that per slide.
- Progress indicator as a segmented bar at the top (Instagram-story convention —
  the one "expected pattern" worth keeping, because breaking it costs usability for
  no gain).
- Export button on the last slide only: renders the current slide to PNG via canvas.

### Widget

- Deliberately minimal — it has to look right embedded in someone else's README on
  GitHub's white background *and* a personal site's dark background, so it ships as
  an SVG with a transparent background, `signal-coral` accent only, no card/box
  around it.

### Cross-device behavior (PWA)

- `manifest.json` + service worker for installability (Add to Home Screen).
- Mobile-first breakpoints; story view is the primary mobile experience, dashboard
  the primary desktop experience — both fully usable on either, but design each for
  its likeliest device.
- Respect `prefers-reduced-motion` — the waveform signature falls back to a static
  version; this is a portfolio piece, an accessibility miss here reads badly to
  anyone reviewing your code.

---

## 7. Build order (recommended for Claude Code)

1. Postgres schema + migrations
2. Docker Compose (Postgres + api + worker) so the rest of the build runs against
   a consistent local environment from here on
3. Go: OAuth flow + session handling
4. Go: poller (worker mode) writing to `listening_events`
5. Go: stats aggregation endpoints (reading from `top_snapshots`)
6. Go: widget endpoint + caching
7. React: design tokens + shell (Space Grotesk/Inter/JetBrains Mono loaded, color
   variables in place) before any real screens — get the system right once
8. React: login/landing screen (the one that matters most)
9. React: dashboard
10. React: story view + export
11. PWA manifest + service worker
12. Wrap `apps/web` with Capacitor → generate `ios/` and `android/` projects, confirm
    it runs in a simulator/emulator
13. Caddyfile + production compose profile for one-command deploy

---

## 8. Open decisions to confirm before build

- Redis for widget caching, or in-memory `sync.Map` (simpler, fine for 5-user dev-mode cap)?
- Session strategy: signed cookie (simpler) vs JWT in an Authorization header (more
  portfolio-visible "I understand auth" signal)?
- Poller as a goroutine inside the `api` binary for v1, or genuinely separate
  `worker` deployable from day one?
