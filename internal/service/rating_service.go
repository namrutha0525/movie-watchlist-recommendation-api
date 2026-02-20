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

type RatingService struct {
	ratingRepo   repository.RatingRepository
	movieService *MovieService
	logger       *zap.Logger
}

func NewRatingService(
	ratingRepo repository.RatingRepository,
	movieService *MovieService,
	logger *zap.Logger,
) *RatingService {
	return &RatingService{
		ratingRepo:   ratingRepo,
		movieService: movieService,
		logger:       logger,
	}
}

// Create rates a movie. The movie is fetched/created from OMDb if not in DB.
func (s *RatingService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateRatingRequest) (*domain.Rating, error) {
	movie, err := s.movieService.GetByImdbID(ctx, req.ImdbID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rating := &domain.Rating{
		ID:        uuid.New(),
		UserID:    userID,
		MovieID:   movie.ID,
		Score:     req.Score,
		Review:    req.Review,
		CreatedAt: now,
		UpdatedAt: now,
		Movie:     movie,
	}

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		if errors.Is(err, appErr.ErrAlreadyExists) {
			return nil, appErr.New(409, "you have already rated this movie", appErr.ErrAlreadyExists)
		}
		s.logger.Error("failed to create rating", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	return rating, nil
}

// GetAll returns all ratings for the user.
func (s *RatingService) GetAll(ctx context.Context, userID uuid.UUID) ([]domain.Rating, error) {
	ratings, err := s.ratingRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get ratings", zap.Error(err))
		return nil, appErr.ErrInternal
	}
	return ratings, nil
}

// Update modifies an existing rating.
func (s *RatingService) Update(ctx context.Context, userID uuid.UUID, ratingID uuid.UUID, req *domain.UpdateRatingRequest) (*domain.Rating, error) {
	rating, err := s.ratingRepo.GetByID(ctx, ratingID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, appErr.ErrNotFound
		}
		return nil, appErr.ErrInternal
	}
	if rating.UserID != userID {
		return nil, appErr.ErrForbidden
	}

	rating.Score = req.Score
	rating.Review = req.Review
	rating.UpdatedAt = time.Now()

	if err := s.ratingRepo.Update(ctx, rating); err != nil {
		s.logger.Error("failed to update rating", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	return rating, nil
}

// Delete removes a rating.
func (s *RatingService) Delete(ctx context.Context, userID uuid.UUID, ratingID uuid.UUID) error {
	rating, err := s.ratingRepo.GetByID(ctx, ratingID)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return appErr.ErrNotFound
		}
		return appErr.ErrInternal
	}
	if rating.UserID != userID {
		return appErr.ErrForbidden
	}

	return s.ratingRepo.Delete(ctx, ratingID)
}
