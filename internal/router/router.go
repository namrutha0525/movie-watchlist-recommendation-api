package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/namru/movie-recommend/internal/handler"
	"github.com/namru/movie-recommend/internal/middleware"
)

// Setup initializes the Gin router with all routes and middleware.
func Setup(
	logger *zap.Logger,
	jwtSecret string,
	authHandler *handler.AuthHandler,
	movieHandler *handler.MovieHandler,
	watchlistHandler *handler.WatchlistHandler,
	ratingHandler *handler.RatingHandler,
	recHandler *handler.RecommendationHandler,
) *gin.Engine {
	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))

	// Rate limiter: 100 requests per minute per IP
	limiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(limiter.Middleware())

	// Health check (no auth)
	r.GET("/api/v1/health", middleware.HealthCheck)

	// Auth routes (no auth required)
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// Movies
		protected.GET("/movies/search", movieHandler.Search)
		protected.GET("/movies/:imdbID", movieHandler.GetByImdbID)

		// Watchlist
		protected.GET("/watchlist", watchlistHandler.GetAll)
		protected.POST("/watchlist", watchlistHandler.Add)
		protected.PATCH("/watchlist/:id", watchlistHandler.UpdateStatus)
		protected.DELETE("/watchlist/:id", watchlistHandler.Remove)

		// Ratings
		protected.POST("/ratings", ratingHandler.Create)
		protected.GET("/ratings", ratingHandler.GetAll)
		protected.PUT("/ratings/:id", ratingHandler.Update)
		protected.DELETE("/ratings/:id", ratingHandler.Delete)

		// Recommendations
		protected.GET("/recommendations", recHandler.GetRecommendations)
	}

	return r
}
