package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gruvit/server/go/models"
	"gruvit/server/go/services"
)

type SearchHandler struct {
	externalAPIService *services.ExternalAPIService
	redisService       *services.RedisService
}

func NewSearchHandler(externalAPIService *services.ExternalAPIService, redisService *services.RedisService) *SearchHandler {
	return &SearchHandler{
		externalAPIService: externalAPIService,
		redisService:       redisService,
	}
}

func (h *SearchHandler) SearchTracks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Check cache first
	cachedResponse, err := h.redisService.GetSearchResults(query)
	if err == nil && cachedResponse != nil {
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// Search external APIs
	tracks, err := h.externalAPIService.SearchTracks(query, limit)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search tracks", "details": err.Error()})
		return
	}

	// Create response
	response := &models.SearchResponse{
		Query:   query,
		Results: tracks,
		Total:   len(tracks),
		Page:    page,
		Limit:   limit,
	}

	// Cache the results for 1 hour
	h.redisService.SetSearchResults(query, response, time.Hour)

	c.JSON(http.StatusOK, response)
}
