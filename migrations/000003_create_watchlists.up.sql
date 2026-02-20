CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'plan_to_watch',
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_watchlist_user_movie UNIQUE (user_id, movie_id),
    CONSTRAINT chk_watchlist_status CHECK (status IN ('plan_to_watch', 'watching', 'watched'))
);

CREATE INDEX idx_watchlists_user_id ON watchlists(user_id);
