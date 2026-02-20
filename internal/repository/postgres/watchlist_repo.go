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

type WatchlistRepo struct {
	pool *pgxpool.Pool
}

func NewWatchlistRepo(pool *pgxpool.Pool) *WatchlistRepo {
	return &WatchlistRepo{pool: pool}
}

func (r *WatchlistRepo) Create(ctx context.Context, entry *domain.Watchlist) error {
	query := `
		INSERT INTO watchlists (id, user_id, movie_id, status, added_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.UserID, entry.MovieID, entry.Status, entry.AddedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return appErr.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *WatchlistRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Watchlist, error) {
	query := `
		SELECT w.id, w.user_id, w.movie_id, w.status, w.added_at,
		       m.id, m.imdb_id, m.title, m.year, m.genre, m.director, m.actors, m.plot, m.poster_url, m.imdb_rating, m.created_at
		FROM watchlists w
		JOIN movies m ON m.id = w.movie_id
		WHERE w.id = $1`

	var w domain.Watchlist
	var m domain.Movie
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.UserID, &w.MovieID, &w.Status, &w.AddedAt,
		&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Genre, &m.Director,
		&m.Actors, &m.Plot, &m.PosterURL, &m.ImdbRating, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.ErrNotFound
		}
		return nil, err
	}
	w.Movie = &m
	return &w, nil
}

func (r *WatchlistRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Watchlist, error) {
	query := `
		SELECT w.id, w.user_id, w.movie_id, w.status, w.added_at,
		       m.id, m.imdb_id, m.title, m.year, m.genre, m.director, m.actors, m.plot, m.poster_url, m.imdb_rating, m.created_at
		FROM watchlists w
		JOIN movies m ON m.id = w.movie_id
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Watchlist
	for rows.Next() {
		var w domain.Watchlist
		var m domain.Movie
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.MovieID, &w.Status, &w.AddedAt,
			&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Genre, &m.Director,
			&m.Actors, &m.Plot, &m.PosterURL, &m.ImdbRating, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		w.Movie = &m
		list = append(list, w)
	}
	return list, rows.Err()
}

func (r *WatchlistRepo) Update(ctx context.Context, id uuid.UUID, status domain.WatchlistStatus) error {
	query := `UPDATE watchlists SET status = $1 WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrNotFound
	}
	return nil
}

func (r *WatchlistRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM watchlists WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrNotFound
	}
	return nil
}

func (r *WatchlistRepo) Exists(ctx context.Context, userID, movieID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM watchlists WHERE user_id = $1 AND movie_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, movieID).Scan(&exists)
	return exists, err
}
