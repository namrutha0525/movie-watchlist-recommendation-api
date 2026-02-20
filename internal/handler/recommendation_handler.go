package handler

import (
	"github.com/gin-gonic/gin"

	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/service"
	"github.com/namru/movie-recommend/pkg/response"
)

type RecommendationHandler struct {
	recService *service.RecommendationService
}

func NewRecommendationHandler(recService *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: recService}
}

// GetRecommendations returns personalized movie recommendations.
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	userID := getUserID(c)

	movies, err := h.recService.GetRecommendations(c.Request.Context(), userID)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "recommendations generated", movies)
}
