package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/namru/movie-recommend/internal/config"
	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/repository"
)

type MovieService struct {
	movieRepo repository.MovieRepository
	cache     repository.CacheRepository
	cfg       *config.Config
	logger    *zap.Logger
	client    *http.Client
}

func NewMovieService(
	movieRepo repository.MovieRepository,
	cache repository.CacheRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *MovieService {
	return &MovieService{
		movieRepo: movieRepo,
		cache:     cache,
		cfg:       cfg,
		logger:    logger,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Search queries OMDb for movies by title (with Redis caching).
func (s *MovieService) Search(ctx context.Context, query string, page int) (*domain.OMDbSearchResponse, error) {
	if page <= 0 {
		page = 1
	}

	// Check cache
	cacheKey := fmt.Sprintf("omdb:search:%s:%d", query, page)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("cache get error", zap.Error(err))
	}
	if cached != "" {
		var result domain.OMDbSearchResponse
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			s.logger.Debug("cache hit", zap.String("key", cacheKey))
			return &result, nil
		}
	}

	// Call OMDb API
	url := fmt.Sprintf("%s/?apikey=%s&s=%s&page=%d",
		s.cfg.OMDB.BaseURL, s.cfg.OMDB.APIKey, query, page,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		s.logger.Error("omdb api call failed", zap.Error(err))
		return nil, appErr.ErrExternalAPI
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read omdb response", zap.Error(err))
		return nil, appErr.ErrExternalAPI
	}

	var result domain.OMDbSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		s.logger.Error("failed to parse omdb response", zap.Error(err))
		return nil, appErr.ErrExternalAPI
	}

	if result.Response == "False" {
		return nil, appErr.New(404, "no movies found: "+result.Error, appErr.ErrNotFound)
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, string(body), s.cfg.Cache.SearchTTL); err != nil {
		s.logger.Warn("cache set error", zap.Error(err))
	}

	return &result, nil
}

// GetByImdbID fetches full movie details from OMDb (with caching) and persists to DB.
func (s *MovieService) GetByImdbID(ctx context.Context, imdbID string) (*domain.Movie, error) {
	// Check DB first
	movie, err := s.movieRepo.GetByImdbID(ctx, imdbID)
	if err == nil {
		return movie, nil
	}

	// Check cache
	cacheKey := fmt.Sprintf("omdb:movie:%s", imdbID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		s.logger.Warn("cache get error", zap.Error(err))
	}

	var detail domain.OMDbMovieDetail
	if cached != "" {
		if err := json.Unmarshal([]byte(cached), &detail); err == nil {
			s.logger.Debug("cache hit for movie detail", zap.String("imdbID", imdbID))
			return s.persistMovie(&detail)
		}
	}

	// Call OMDb API
	url := fmt.Sprintf("%s/?apikey=%s&i=%s&plot=full",
		s.cfg.OMDB.BaseURL, s.cfg.OMDB.APIKey, imdbID,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		s.logger.Error("omdb api call failed", zap.Error(err))
		return nil, appErr.ErrExternalAPI
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, appErr.ErrExternalAPI
	}

	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, appErr.ErrExternalAPI
	}

	if detail.Response == "False" {
		return nil, appErr.New(404, "movie not found: "+detail.Error, appErr.ErrNotFound)
	}

	// Cache
	if err := s.cache.Set(ctx, cacheKey, string(body), s.cfg.Cache.MovieTTL); err != nil {
		s.logger.Warn("cache set error", zap.Error(err))
	}

	return s.persistMovie(&detail)
}

// persistMovie saves the OMDb movie detail to the database.
func (s *MovieService) persistMovie(detail *domain.OMDbMovieDetail) (*domain.Movie, error) {
	movie := &domain.Movie{
		ID:         uuid.New(),
		ImdbID:     detail.ImdbID,
		Title:      detail.Title,
		Year:       detail.Year,
		Genre:      detail.Genre,
		Director:   detail.Director,
		Actors:     detail.Actors,
		Plot:       detail.Plot,
		PosterURL:  detail.Poster,
		ImdbRating: detail.ImdbRating,
		CreatedAt:  time.Now(),
	}

	if err := s.movieRepo.Create(context.Background(), movie); err != nil {
		s.logger.Warn("failed to persist movie (may already exist)", zap.Error(err))
		// Try to return the existing movie from DB
		existing, dbErr := s.movieRepo.GetByImdbID(context.Background(), detail.ImdbID)
		if dbErr == nil {
			return existing, nil
		}
	}

	return movie, nil
}
