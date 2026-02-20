package handler

import (
	"github.com/gin-gonic/gin"

	appErr "github.com/namru/movie-recommend/internal/errors"
	"github.com/namru/movie-recommend/internal/service"
	"github.com/namru/movie-recommend/pkg/response"
)

type MovieHandler struct {
	movieService *service.MovieService
}

func NewMovieHandler(movieService *service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

// Search godoc
// @Summary Search movies via OMDb
// @Tags movies
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Router /api/v1/movies/search [get]
func (h *MovieHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "query parameter 'q' is required")
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		// Simple conversion
		for _, ch := range p {
			if ch >= '0' && ch <= '9' {
				page = page*10 + int(ch-'0')
			}
		}
		if page <= 0 {
			page = 1
		}
	}

	result, err := h.movieService.Search(c.Request.Context(), query, page)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "movies found", result)
}

// GetByImdbID godoc
// @Summary Get movie details by IMDb ID
// @Tags movies
// @Produce json
// @Param imdbID path string true "IMDb ID (e.g. tt1234567)"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Router /api/v1/movies/{imdbID} [get]
func (h *MovieHandler) GetByImdbID(c *gin.Context) {
	imdbID := c.Param("imdbID")
	if imdbID == "" {
		response.BadRequest(c, "imdbID parameter is required")
		return
	}

	movie, err := h.movieService.GetByImdbID(c.Request.Context(), imdbID)
	if err != nil {
		status := appErr.MapToHTTPStatus(err)
		c.JSON(status, response.APIResponse{Success: false, Error: err.Error()})
		return
	}

	response.OK(c, "movie details retrieved", movie)
}
