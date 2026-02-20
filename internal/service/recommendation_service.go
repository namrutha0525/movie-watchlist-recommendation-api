package service

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/repository"
)

type RecommendationService struct {
	ratingRepo   repository.RatingRepository
	movieRepo    repository.MovieRepository
	movieService *MovieService
	logger       *zap.Logger
}

func NewRecommendationService(
	ratingRepo repository.RatingRepository,
	movieRepo repository.MovieRepository,
	movieService *MovieService,
	logger *zap.Logger,
) *RecommendationService {
	return &RecommendationService{
		ratingRepo:   ratingRepo,
		movieRepo:    movieRepo,
		movieService: movieService,
		logger:       logger,
	}
}

// GetRecommendations returns personalized movie recommendations for the user.
//
// Algorithm:
// 1. Find the user's top genres from highly-rated movies (score >= 7).
// 2. Search OMDb for movies in those genres.
// 3. Exclude movies the user has already rated or added to watchlist.
// 4. Return up to 10 recommendations.
func (s *RecommendationService) GetRecommendations(ctx context.Context, userID uuid.UUID) ([]domain.Movie, error) {
	// Step 1: Get top genres from user's highly-rated movies
	genres, err := s.ratingRepo.GetTopGenresByUser(ctx, userID, 7, 3)
	if err != nil {
		s.logger.Error("failed to get top genres", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	if len(genres) == 0 {
		// No highly-rated movies — return a default set
		genres = []string{"Action", "Drama", "Comedy"}
		s.logger.Info("no rated movies found, using default genres")
	}

	// Step 2: Get already-rated movie IDs for exclusion
	ratedIDs, err := s.ratingRepo.GetRatedMovieIDs(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get rated movie IDs", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	ratedSet := make(map[uuid.UUID]bool)
	for _, id := range ratedIDs {
		ratedSet[id] = true
	}

	// Step 3: Search for movies by top genres (from local DB first, then OMDb)
	var recommendations []domain.Movie

	for _, genre := range genres {
		// Search local DB
		movies, err := s.movieRepo.GetByGenre(ctx, strings.TrimSpace(genre), 20)
		if err != nil {
			s.logger.Warn("failed to search movies by genre", zap.String("genre", genre), zap.Error(err))
			continue
		}

		for _, m := range movies {
			if ratedSet[m.ID] {
				continue
			}
			recommendations = append(recommendations, m)
			ratedSet[m.ID] = true // prevent duplicates across genres

			if len(recommendations) >= 10 {
				return recommendations, nil
			}
		}
	}

	// Step 4: If local DB didn't yield enough, try OMDb search for each genre
	if len(recommendations) < 10 {
		for _, genre := range genres {
			searchResult, err := s.movieService.Search(ctx, strings.TrimSpace(genre), 1)
			if err != nil {
				s.logger.Warn("omdb genre search failed", zap.String("genre", genre), zap.Error(err))
				continue
			}

			for _, sr := range searchResult.Search {
				movie, err := s.movieService.GetByImdbID(ctx, sr.ImdbID)
				if err != nil {
					continue
				}
				if ratedSet[movie.ID] {
					continue
				}

				recommendations = append(recommendations, *movie)
				ratedSet[movie.ID] = true

				if len(recommendations) >= 10 {
					return recommendations, nil
				}
			}
		}
	}

	return recommendations, nil
}
