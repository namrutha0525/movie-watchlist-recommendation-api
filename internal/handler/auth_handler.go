package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/service"
	"github.com/namru/movie-recommend/pkg/response"
	"github.com/namru/movie-recommend/pkg/validator"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.RegisterRequest true "Register"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		errors := validator.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Error:   "validation failed",
			Data:    errors,
		})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.Created(c, "user registered successfully", result)
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Login"
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		errors := validator.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, response.APIResponse{
			Success: false,
			Error:   "validation failed",
			Data:    errors,
		})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "login successful", result)
}
