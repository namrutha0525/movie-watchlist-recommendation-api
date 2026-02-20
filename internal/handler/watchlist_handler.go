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

type WatchlistHandler struct {
	watchlistService *service.WatchlistService
}

func NewWatchlistHandler(watchlistService *service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistService: watchlistService}
}

// GetAll returns the user's watchlist.
func (h *WatchlistHandler) GetAll(c *gin.Context) {
	userID := getUserID(c)

	entries, err := h.watchlistService.GetAll(c.Request.Context(), userID)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "watchlist retrieved", entries)
}

// Add adds a movie to the watchlist.
func (h *WatchlistHandler) Add(c *gin.Context) {
	userID := getUserID(c)

	var req domain.AddToWatchlistRequest
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

	entry, err := h.watchlistService.Add(c.Request.Context(), userID, &req)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.Created(c, "movie added to watchlist", entry)
}

// UpdateStatus updates a watchlist entry's status.
func (h *WatchlistHandler) UpdateStatus(c *gin.Context) {
	userID := getUserID(c)

	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid watchlist entry ID")
		return
	}

	var req domain.UpdateWatchlistRequest
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

	if err := h.watchlistService.UpdateStatus(c.Request.Context(), userID, entryID, &req); err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "watchlist entry updated", nil)
}

// Remove deletes a watchlist entry.
func (h *WatchlistHandler) Remove(c *gin.Context) {
	userID := getUserID(c)

	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid watchlist entry ID")
		return
	}

	if err := h.watchlistService.Remove(c.Request.Context(), userID, entryID); err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "movie removed from watchlist", nil)
}

// getUserID extracts the user ID set by the auth middleware.
func getUserID(c *gin.Context) uuid.UUID {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	return userID
}
