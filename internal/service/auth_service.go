package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/namru/movie-recommend/internal/config"
	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/repository"
)

type AuthService struct {
	userRepo repository.UserRepository
	cfg      *config.JWTConfig
	logger   *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.JWTConfig, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, appErr.ErrAlreadyExists) {
			return nil, appErr.New(409, "username or email already exists", appErr.ErrAlreadyExists)
		}
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	// Generate JWT
	token, err := s.generateToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

// Login authenticates a user and returns a JWT.
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, appErr.ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, appErr.ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return nil, appErr.ErrInternal
	}

	return &domain.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Duration(s.cfg.ExpiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}
