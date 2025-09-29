package handlers

import (
	"net/http"
	"strconv"

	"gruvit/server/go/services"

	"github.com/gin-gonic/gin"
)

type SocialHandler struct {
	socialService *services.SocialService
}

func NewSocialHandler(socialService *services.SocialService) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

// GetUserFeed returns the social feed for a user
func (h *SocialHandler) GetUserFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	activities, err := h.socialService.GetUserFeed(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get social feed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// GetUserActivities returns activities for a specific user
func (h *SocialHandler) GetUserActivities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	activities, err := h.socialService.GetUserActivities(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

// FollowUser makes a user follow another user
func (h *SocialHandler) FollowUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user ID is required"})
		return
	}

	if userID.(string) == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
		return
	}

	err := h.socialService.FollowUser(userID.(string), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}

// UnfollowUser makes a user unfollow another user
func (h *SocialHandler) UnfollowUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user ID is required"})
		return
	}

	err := h.socialService.UnfollowUser(userID.(string), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}

// GetUserStats returns social statistics for a user
func (h *SocialHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	stats, err := h.socialService.GetUserStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetNotifications returns notifications for a user
func (h *SocialHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	notifications, err := h.socialService.GetNotifications(userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

// MarkNotificationAsRead marks a notification as read
func (h *SocialHandler) MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("notificationId")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	err := h.socialService.MarkNotificationAsRead(notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// RecordActivity records a social activity
func (h *SocialHandler) RecordActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		ActivityType string                 `json:"activity_type" binding:"required"`
		TargetID     string                 `json:"target_id" binding:"required"`
		TargetType   string                 `json:"target_type" binding:"required"`
		Metadata     map[string]interface{} `json:"metadata"`
		IsPublic     bool                   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.socialService.RecordActivity(
		userID.(string),
		request.ActivityType,
		request.TargetID,
		request.TargetType,
		request.Metadata,
		request.IsPublic,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity recorded successfully"})
}
