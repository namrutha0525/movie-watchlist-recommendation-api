package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
)

type MovieRepo struct {
	pool *pgxpool.Pool
}

func NewMovieRepo(pool *pgxpool.Pool) *MovieRepo {
	return &MovieRepo{pool: pool}
}

func (r *MovieRepo) Create(ctx context.Context, movie *domain.Movie) error {
	query := `
		INSERT INTO movies (id, imdb_id, title, year, genre, director, actors, plot, poster_url, imdb_rating, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (imdb_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query,
		movie.ID, movie.ImdbID, movie.Title, movie.Year, movie.Genre,
		movie.Director, movie.Actors, movie.Plot, movie.PosterURL,
		movie.ImdbRating, movie.CreatedAt,
	)
	return err
}

func (r *MovieRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Movie, error) {
	query := `SELECT id, imdb_id, title, year, genre, director, actors, plot, poster_url, imdb_rating, created_at
	           FROM movies WHERE id = $1`

	var movie domain.Movie
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&movie.ID, &movie.ImdbID, &movie.Title, &movie.Year, &movie.Genre,
		&movie.Director, &movie.Actors, &movie.Plot, &movie.PosterURL,
		&movie.ImdbRating, &movie.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.ErrNotFound
		}
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepo) GetByImdbID(ctx context.Context, imdbID string) (*domain.Movie, error) {
	query := `SELECT id, imdb_id, title, year, genre, director, actors, plot, poster_url, imdb_rating, created_at
	           FROM movies WHERE imdb_id = $1`

	var movie domain.Movie
	err := r.pool.QueryRow(ctx, query, imdbID).Scan(
		&movie.ID, &movie.ImdbID, &movie.Title, &movie.Year, &movie.Genre,
		&movie.Director, &movie.Actors, &movie.Plot, &movie.PosterURL,
		&movie.ImdbRating, &movie.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.ErrNotFound
		}
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepo) GetByGenre(ctx context.Context, genre string, limit int) ([]domain.Movie, error) {
	query := `SELECT id, imdb_id, title, year, genre, director, actors, plot, poster_url, imdb_rating, created_at
	           FROM movies WHERE genre ILIKE '%' || $1 || '%' LIMIT $2`

	rows, err := r.pool.Query(ctx, query, genre, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []domain.Movie
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(
			&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Genre,
			&m.Director, &m.Actors, &m.Plot, &m.PosterURL,
			&m.ImdbRating, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, rows.Err()
}
