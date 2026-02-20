-- =============================================================
-- Movie Watchlist & Recommendation API — Full PostgreSQL Schema
-- =============================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================
-- 1. USERS TABLE
-- =============================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email    UNIQUE (email)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- =============================================================
-- 2. MOVIES TABLE (cached from OMDb API)
-- =============================================================
CREATE TABLE IF NOT EXISTS movies (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    imdb_id     VARCHAR(20)  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    year        VARCHAR(10),
    genre       VARCHAR(255),
    director    VARCHAR(255),
    actors      TEXT,
    plot        TEXT,
    poster_url  TEXT,
    imdb_rating VARCHAR(10),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_movies_imdb_id UNIQUE (imdb_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_movies_imdb_id ON movies(imdb_id);
CREATE INDEX IF NOT EXISTS idx_movies_genre   ON movies(genre);
CREATE INDEX IF NOT EXISTS idx_movies_title   ON movies(title);

-- =============================================================
-- 3. WATCHLISTS TABLE
-- =============================================================
CREATE TABLE IF NOT EXISTS watchlists (
    id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  UUID        NOT NULL,
    movie_id UUID        NOT NULL,
    status   VARCHAR(20) NOT NULL DEFAULT 'plan_to_watch',
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign Keys
    CONSTRAINT fk_watchlists_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_watchlists_movie
        FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,

    -- One watchlist entry per movie per user
    CONSTRAINT uq_watchlist_user_movie UNIQUE (user_id, movie_id),

    -- Status must be one of the allowed values
    CONSTRAINT chk_watchlist_status
        CHECK (status IN ('plan_to_watch', 'watching', 'watched'))
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_watchlists_user_id  ON watchlists(user_id);
CREATE INDEX IF NOT EXISTS idx_watchlists_movie_id ON watchlists(movie_id);
CREATE INDEX IF NOT EXISTS idx_watchlists_status   ON watchlists(status);

-- =============================================================
-- 4. RATINGS TABLE
-- =============================================================
CREATE TABLE IF NOT EXISTS ratings (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    movie_id   UUID        NOT NULL,
    score      INTEGER     NOT NULL,
    review     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign Keys
    CONSTRAINT fk_ratings_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_ratings_movie
        FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,

    -- One rating per movie per user
    CONSTRAINT uq_rating_user_movie UNIQUE (user_id, movie_id),

    -- Score must be between 1 and 10
    CONSTRAINT chk_rating_score CHECK (score >= 1 AND score <= 10)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_ratings_user_id  ON ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_ratings_movie_id ON ratings(movie_id);
CREATE INDEX IF NOT EXISTS idx_ratings_score    ON ratings(score);

-- =============================================================
-- 5. AUTO-UPDATE updated_at TRIGGER
-- =============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to users
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to ratings
CREATE TRIGGER trg_ratings_updated_at
    BEFORE UPDATE ON ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
