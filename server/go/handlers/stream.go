package handlers

import (
	"net/http"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type StreamHandler struct {
	streamingService *services.StreamingService
}

func NewStreamHandler(streamingService *services.StreamingService) *StreamHandler {
	return &StreamHandler{
		streamingService: streamingService,
	}
}

func (h *StreamHandler) GetStreamURL(c *gin.Context) {
	trackID := c.Param("trackId")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
		return
	}

	source := c.Query("source")
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source parameter is required"})
		return
	}

	// Use the streaming service to get the stream URL
	response, err := h.streamingService.GetStreamURL(trackID, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return in the format expected by frontend: { stream_url: string }
	c.JSON(http.StatusOK, gin.H{
		"stream_url": response.StreamURL,
	})
}
