package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/namru/movie-recommend/internal/domain"
	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/service"
	"github.com/namru/movie-recommend/pkg/response"
	"github.com/namru/movie-recommend/pkg/validator"
)

type RatingHandler struct {
	ratingService *service.RatingService
}

func NewRatingHandler(ratingService *service.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

// Create rates a movie.
func (h *RatingHandler) Create(c *gin.Context) {
	userID := getUserID(c)

	var req domain.CreateRatingRequest
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

	rating, err := h.ratingService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.Created(c, "movie rated successfully", rating)
}

// GetAll returns all ratings for the current user.
func (h *RatingHandler) GetAll(c *gin.Context) {
	userID := getUserID(c)

	ratings, err := h.ratingService.GetAll(c.Request.Context(), userID)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "ratings retrieved", ratings)
}

// Update modifies a rating.
func (h *RatingHandler) Update(c *gin.Context) {
	userID := getUserID(c)

	ratingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid rating ID")
		return
	}

	var req domain.UpdateRatingRequest
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

	rating, err := h.ratingService.Update(c.Request.Context(), userID, ratingID, &req)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "rating updated", rating)
}

// Delete removes a rating.
func (h *RatingHandler) Delete(c *gin.Context) {
	userID := getUserID(c)

	ratingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid rating ID")
		return
	}

	if err := h.ratingService.Delete(c.Request.Context(), userID, ratingID); err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "rating deleted", nil)
}
