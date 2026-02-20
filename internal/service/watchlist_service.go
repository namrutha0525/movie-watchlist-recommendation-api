package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/repository"
)

type WatchlistService struct {
	watchlistRepo repository.WatchlistRepository
	movieService  *MovieService
	logger        *zap.Logger
}

func NewWatchlistService(
	watchlistRepo repository.WatchlistRepository,
	movieService *MovieService,
	logger *zap.Logger,
) *WatchlistService {
	return &WatchlistService{
		watchlistRepo: watchlistRepo,
		movieService:  movieService,
		logger:        logger,
	}
}

// Add adds a movie to the user's watchlist. Fetches the movie from OMDb if not in DB.
func (s *WatchlistService) Add(ctx context.Context, userID uuid.UUID, req *domain.AddToWatchlistRequest) (*domain.Watchlist, error) {
	// Fetch or create the movie
	movie, err := s.movieService.GetByImdbID(ctx, req.ImdbID)
	if err != nil {
		return nil, err
	}

	// Check if already in watchlist
	exists, err := s.watchlistRepo.Exists(ctx, userID, movie.ID)
	if err != nil {
		s.logger.Error("failed to check watchlist existence", zap.Error(err))
		return nil, appErr.ErrInternal
	}
	if exists {
		return nil, appErr.New(409, "movie already in watchlist", appErr.ErrAlreadyExists)
	}

	status := req.Status
	if status == "" {
		status = domain.StatusPlanToWatch
	}

	entry := &domain.Watchlist{
		ID:      uuid.New(),
		UserID:  userID,
		MovieID: movie.ID,
		Status:  status,
		AddedAt: time.Now(),
		Movie:   movie,
	}

	if err := s.watchlistRepo.Create(ctx, entry); err != nil {
		if errors.Is(err, appErr.ErrAlreadyExists) {
			return nil, appErr.New(409, "movie already in watchlist", appErr.ErrAlreadyExists)
		}
		s.logger.Error("failed to add to watchlist", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	return entry, nil
}

// GetAll returns all watchlist entries for the user.
func (s *WatchlistService) GetAll(ctx context.Context, userID uuid.UUID) ([]domain.Watchlist, error) {
	entries, err := s.watchlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get watchlist", zap.Error(err))
		return nil, appErr.ErrInternal
	}
	return entries, nil
}

// UpdateStatus updates a watchlist entry's status.
func (s *WatchlistService) UpdateStatus(ctx context.Context, userID uuid.UUID, entryID uuid.UUID, req *domain.UpdateWatchlistRequest) error {
	// Verify ownership
	entry, err := s.watchlistRepo.GetByID(ctx, entryID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return appErr.ErrNotFound
		}
		return appErr.ErrInternal
	}
	if entry.UserID != userID {
		return appErr.ErrForbidden
	}

	return s.watchlistRepo.Update(ctx, entryID, req.Status)
}

// Remove deletes a watchlist entry.
func (s *WatchlistService) Remove(ctx context.Context, userID uuid.UUID, entryID uuid.UUID) error {
	entry, err := s.watchlistRepo.GetByID(ctx, entryID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return appErr.ErrNotFound
		}
		return appErr.ErrInternal
	}
	if entry.UserID != userID {
		return appErr.ErrForbidden
	}

	return s.watchlistRepo.Delete(ctx, entryID)
}
