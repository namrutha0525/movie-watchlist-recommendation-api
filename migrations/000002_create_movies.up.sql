CREATE TABLE movies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    imdb_id VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    year VARCHAR(10),
    genre VARCHAR(255),
    director VARCHAR(255),
    actors TEXT,
    plot TEXT,
    poster_url TEXT,
    imdb_rating VARCHAR(10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_movies_imdb_id ON movies(imdb_id);
CREATE INDEX idx_movies_genre ON movies(genre);
