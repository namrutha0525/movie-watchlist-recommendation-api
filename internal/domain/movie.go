package domain

import (
	"time"

	"github.com/google/uuid"
)

// Movie represents a movie (cached from OMDb).
type Movie struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ImdbID     string    `json:"imdb_id" db:"imdb_id"`
	Title      string    `json:"title" db:"title"`
	Year       string    `json:"year" db:"year"`
	Genre      string    `json:"genre" db:"genre"`
	Director   string    `json:"director" db:"director"`
	Actors     string    `json:"actors" db:"actors"`
	Plot       string    `json:"plot" db:"plot"`
	PosterURL  string    `json:"poster_url" db:"poster_url"`
	ImdbRating string    `json:"imdb_rating" db:"imdb_rating"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// OMDbSearchResult represents a single item from OMDb search.
type OMDbSearchResult struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

// OMDbSearchResponse is the full OMDb search API response.
type OMDbSearchResponse struct {
	Search       []OMDbSearchResult `json:"Search"`
	TotalResults string             `json:"totalResults"`
	Response     string             `json:"Response"`
	Error        string             `json:"Error"`
}

// OMDbMovieDetail is the detailed OMDb movie response.
type OMDbMovieDetail struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Rated      string `json:"Rated"`
	Released   string `json:"Released"`
	Runtime    string `json:"Runtime"`
	Genre      string `json:"Genre"`
	Director   string `json:"Director"`
	Writer     string `json:"Writer"`
	Actors     string `json:"Actors"`
	Plot       string `json:"Plot"`
	Language   string `json:"Language"`
	Country    string `json:"Country"`
	Awards     string `json:"Awards"`
	Poster     string `json:"Poster"`
	ImdbRating string `json:"imdbRating"`
	ImdbID     string `json:"imdbID"`
	Type       string `json:"Type"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}
