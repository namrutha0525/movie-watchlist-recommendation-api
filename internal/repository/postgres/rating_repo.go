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

type RatingRepo struct {
	pool *pgxpool.Pool
}

func NewRatingRepo(pool *pgxpool.Pool) *RatingRepo {
	return &RatingRepo{pool: pool}
}

func (r *RatingRepo) Create(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		rating.ID, rating.UserID, rating.MovieID, rating.Score,
		rating.Review, rating.CreatedAt, rating.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return appErr.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *RatingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error) {
	query := `
		SELECT r.id, r.user_id, r.movie_id, r.score, r.review, r.created_at, r.updated_at,
		       m.id, m.imdb_id, m.title, m.year, m.genre, m.director, m.actors, m.plot, m.poster_url, m.imdb_rating, m.created_at
		FROM ratings r
		JOIN movies m ON m.id = r.movie_id
		WHERE r.id = $1`

	var rt domain.Rating
	var m domain.Movie
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.UserID, &rt.MovieID, &rt.Score, &rt.Review,
		&rt.CreatedAt, &rt.UpdatedAt,
		&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Genre, &m.Director,
		&m.Actors, &m.Plot, &m.PosterURL, &m.ImdbRating, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErr.ErrNotFound
		}
		return nil, err
	}
	rt.Movie = &m
	return &rt, nil
}

func (r *RatingRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Rating, error) {
	query := `
		SELECT r.id, r.user_id, r.movie_id, r.score, r.review, r.created_at, r.updated_at,
		       m.id, m.imdb_id, m.title, m.year, m.genre, m.director, m.actors, m.plot, m.poster_url, m.imdb_rating, m.created_at
		FROM ratings r
		JOIN movies m ON m.id = r.movie_id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.Rating
	for rows.Next() {
		var rt domain.Rating
		var m domain.Movie
		if err := rows.Scan(
			&rt.ID, &rt.UserID, &rt.MovieID, &rt.Score, &rt.Review,
			&rt.CreatedAt, &rt.UpdatedAt,
			&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Genre, &m.Director,
			&m.Actors, &m.Plot, &m.PosterURL, &m.ImdbRating, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		rt.Movie = &m
		ratings = append(ratings, rt)
	}
	return ratings, rows.Err()
}

func (r *RatingRepo) Update(ctx context.Context, rating *domain.Rating) error {
	query := `UPDATE ratings SET score = $1, review = $2, updated_at = $3 WHERE id = $4`
	tag, err := r.pool.Exec(ctx, query, rating.Score, rating.Review, rating.UpdatedAt, rating.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrNotFound
	}
	return nil
}

func (r *RatingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ratings WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return appErr.ErrNotFound
	}
	return nil
}

// GetTopGenresByUser returns the most common genres from movies the user rated highly.
func (r *RatingRepo) GetTopGenresByUser(ctx context.Context, userID uuid.UUID, minScore int, limit int) ([]string, error) {
	query := `
		SELECT DISTINCT TRIM(g) as genre
		FROM ratings r
		JOIN movies m ON m.id = r.movie_id,
		LATERAL unnest(string_to_array(m.genre, ',')) AS g
		WHERE r.user_id = $1 AND r.score >= $2
		GROUP BY TRIM(g)
		ORDER BY COUNT(*) DESC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, query, userID, minScore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, rows.Err()
}

// GetRatedMovieIDs returns all movie IDs the user has already rated.
func (r *RatingRepo) GetRatedMovieIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT movie_id FROM ratings WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
