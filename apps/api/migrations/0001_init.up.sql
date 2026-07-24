CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    spotify_id    text NOT NULL UNIQUE,
    display_name  text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE spotify_tokens (
    user_id        uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    access_token   text NOT NULL,
    refresh_token  text NOT NULL,
    expires_at     timestamptz NOT NULL
);

CREATE TABLE listening_events (
    id           bigserial PRIMARY KEY,
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id     text NOT NULL,
    track_name   text NOT NULL,
    artist_id    text NOT NULL,
    artist_name  text NOT NULL,
    played_at    timestamptz NOT NULL,
    source       text NOT NULL CHECK (source IN ('recently_played', 'currently_playing')),
    UNIQUE (user_id, track_id, played_at, source)
);

CREATE INDEX idx_listening_events_user_played_at ON listening_events (user_id, played_at DESC);
CREATE INDEX idx_listening_events_artist ON listening_events (artist_id);

CREATE TABLE artist_genres (
    artist_id   text PRIMARY KEY,
    genres      text[] NOT NULL DEFAULT '{}',
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE top_snapshots (
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period         text NOT NULL CHECK (period IN ('short_term', 'medium_term', 'long_term')),
    snapshot_date  date NOT NULL,
    top_artists    jsonb NOT NULL,
    top_tracks     jsonb NOT NULL,
    PRIMARY KEY (user_id, period, snapshot_date)
);

CREATE TABLE widget_tokens (
    token       text PRIMARY KEY,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  timestamptz NOT NULL DEFAULT now(),
    revoked     boolean NOT NULL DEFAULT false
);

CREATE INDEX idx_widget_tokens_user_active ON widget_tokens (user_id) WHERE NOT revoked;
