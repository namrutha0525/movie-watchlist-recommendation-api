package domain

import (
	"time"

	"github.com/google/uuid"
)

// Rating represents a user's rating for a movie.
type Rating struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	MovieID   uuid.UUID `json:"movie_id" db:"movie_id"`
	Score     int       `json:"score" db:"score"`
	Review    string    `json:"review,omitempty" db:"review"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Movie     *Movie    `json:"movie,omitempty"` // joined data
}

// CreateRatingRequest is the input for rating a movie.
type CreateRatingRequest struct {
	ImdbID string `json:"imdb_id" validate:"required"`
	Score  int    `json:"score" validate:"required,gte=1,lte=10"`
	Review string `json:"review" validate:"omitempty,max=1000"`
}

// UpdateRatingRequest is the input for updating a rating.
type UpdateRatingRequest struct {
	Score  int    `json:"score" validate:"required,gte=1,lte=10"`
	Review string `json:"review" validate:"omitempty,max=1000"`
}
