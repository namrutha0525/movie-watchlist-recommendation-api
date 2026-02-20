package domain

import (
	"time"

	"github.com/google/uuid"
)

// WatchlistStatus enumerates the allowed watchlist states.
type WatchlistStatus string

const (
	StatusPlanToWatch WatchlistStatus = "plan_to_watch"
	StatusWatching    WatchlistStatus = "watching"
	StatusWatched     WatchlistStatus = "watched"
)

// Watchlist represents a user's watchlist entry.
type Watchlist struct {
	ID      uuid.UUID       `json:"id" db:"id"`
	UserID  uuid.UUID       `json:"user_id" db:"user_id"`
	MovieID uuid.UUID       `json:"movie_id" db:"movie_id"`
	Status  WatchlistStatus `json:"status" db:"status"`
	AddedAt time.Time       `json:"added_at" db:"added_at"`
	Movie   *Movie          `json:"movie,omitempty"` // joined data
}

// AddToWatchlistRequest is the input for adding a movie to the watchlist.
type AddToWatchlistRequest struct {
	ImdbID string          `json:"imdb_id" validate:"required"`
	Status WatchlistStatus `json:"status" validate:"omitempty,oneof=plan_to_watch watching watched"`
}

// UpdateWatchlistRequest is the input for updating a watchlist entry.
type UpdateWatchlistRequest struct {
	Status WatchlistStatus `json:"status" validate:"required,oneof=plan_to_watch watching watched"`
}
