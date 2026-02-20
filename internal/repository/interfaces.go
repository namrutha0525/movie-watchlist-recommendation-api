package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/namru/movie-recommend/internal/domain"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

// MovieRepository defines persistence operations for movies.
type MovieRepository interface {
	Create(ctx context.Context, movie *domain.Movie) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Movie, error)
	GetByImdbID(ctx context.Context, imdbID string) (*domain.Movie, error)
	GetByGenre(ctx context.Context, genre string, limit int) ([]domain.Movie, error)
}

// WatchlistRepository defines persistence operations for watchlists.
type WatchlistRepository interface {
	Create(ctx context.Context, entry *domain.Watchlist) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Watchlist, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Watchlist, error)
	Update(ctx context.Context, id uuid.UUID, status domain.WatchlistStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, userID, movieID uuid.UUID) (bool, error)
}

// RatingRepository defines persistence operations for ratings.
type RatingRepository interface {
	Create(ctx context.Context, rating *domain.Rating) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Rating, error)
	Update(ctx context.Context, rating *domain.Rating) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTopGenresByUser(ctx context.Context, userID uuid.UUID, minScore int, limit int) ([]string, error)
	GetRatedMovieIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// CacheRepository defines caching operations.
type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
