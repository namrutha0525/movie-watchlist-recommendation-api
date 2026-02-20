package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	goRedis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/namru/movie-recommend/internal/config"
	"github.com/namru/movie-recommend/internal/handler"
	"github.com/namru/movie-recommend/internal/repository/postgres"
	"github.com/namru/movie-recommend/internal/repository/redis"
	"github.com/namru/movie-recommend/internal/router"
	"github.com/namru/movie-recommend/internal/service"
	"github.com/namru/movie-recommend/pkg/logger"
)

func main() {
	// ---------- Config ----------
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ---------- Logger ----------
	zapLogger := logger.New(cfg.Server.GinMode)
	defer zapLogger.Sync()

	// ---------- Gin Mode ----------
	gin.SetMode(cfg.Server.GinMode)

	// ---------- PostgreSQL ----------
	ctx := context.Background()
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		zapLogger.Fatal("failed to parse database config", zap.Error(err))
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		zapLogger.Fatal("failed to ping database", zap.Error(err))
	}
	zapLogger.Info("connected to PostgreSQL", zap.String("host", cfg.Database.Host))

	// ---------- Redis ----------
	rdb := goRedis.NewClient(&goRedis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		zapLogger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer rdb.Close()
	zapLogger.Info("connected to Redis", zap.String("addr", cfg.Redis.Addr()))

	// ---------- Repositories ----------
	userRepo := postgres.NewUserRepo(pool)
	movieRepo := postgres.NewMovieRepo(pool)
	watchlistRepo := postgres.NewWatchlistRepo(pool)
	ratingRepo := postgres.NewRatingRepo(pool)
	cacheRepo := redis.NewCacheRepo(rdb)

	// ---------- Services ----------
	authService := service.NewAuthService(userRepo, &cfg.JWT, zapLogger)
	movieService := service.NewMovieService(movieRepo, cacheRepo, cfg, zapLogger)
	watchlistService := service.NewWatchlistService(watchlistRepo, movieService, zapLogger)
	ratingService := service.NewRatingService(ratingRepo, movieService, zapLogger)
	recService := service.NewRecommendationService(ratingRepo, movieRepo, movieService, zapLogger)

	// ---------- Handlers ----------
	authHandler := handler.NewAuthHandler(authService)
	movieHandler := handler.NewMovieHandler(movieService)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)
	ratingHandler := handler.NewRatingHandler(ratingService)
	recHandler := handler.NewRecommendationHandler(recService)

	// ---------- Router ----------
	r := router.Setup(
		zapLogger,
		cfg.JWT.Secret,
		authHandler,
		movieHandler,
		watchlistHandler,
		ratingHandler,
		recHandler,
	)

	// ---------- Server ----------
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		zapLogger.Info("server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("server stopped gracefully")
}
