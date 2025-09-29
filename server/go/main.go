package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gruvit/server/go/handlers"
	"gruvit/server/go/middleware"
	"gruvit/server/go/models"
	"gruvit/server/go/services"
)

// Global variables for integrated auth
var (
	integratedMongoClient *mongo.Client
	integratedRedisCtx    = context.Background()
	integratedRedisClient *redis.Client
	integratedAuthClient  *services.AuthClient
)

func main() {
	// Load .env
	if err := godotenv.Load("config.dev.env"); err != nil {
		log.Println("No config.dev.env file found")
	}

	// Initialize auth client
	integratedAuthClient = services.NewAuthClient()

	// Check auth service health
	if err := integratedAuthClient.HealthCheck(); err != nil {
		log.Printf("Warning: Auth service health check failed: %v", err)
		log.Println("Continuing without auth service integration...")
	} else {
		log.Println("Successfully connected to auth service")
	}

	// Connect to MongoDB Atlas
	var err error
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}

	integratedMongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	defer integratedMongoClient.Disconnect(context.Background())

	// Test MongoDB connection
	err = integratedMongoClient.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Successfully connected to MongoDB")

	// Connect to Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	integratedRedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	if err := integratedRedisClient.Ping(integratedRedisCtx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Continuing without caching", err)
		integratedRedisClient = nil
	} else {
		log.Println("Successfully connected to Redis")
	}

	// Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Set trusted proxies
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Initialize services
	playlistService := services.NewPlaylistService(integratedMongoClient.Database("gruvit"))
	cacheService := services.NewCacheService(integratedMongoClient.Database("gruvit"))
	userService := services.NewUserService(integratedMongoClient.Database("gruvit"))
	externalAPIService := services.NewExternalAPIService(os.Getenv("JAMENDO_API_KEY"), os.Getenv("JAMENDO_CLIENT_SECRET"))
	streamingService := services.NewStreamingService(integratedRedisClient, externalAPIService)
	webSocketService := services.NewWebSocketService(integratedRedisClient)

	// Initialize rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(integratedRedisClient)

	// Initialize handlers
	authHandler := handlers.NewIntegratedAuthHandler(integratedAuthClient)
	userHandler := handlers.NewUserHandler(playlistService, cacheService, userService)
	streamHandler := handlers.NewStreamHandler(streamingService)
	wsHandler := handlers.NewWebSocketHandler(webSocketService, integratedAuthClient)
	redisService := services.NewRedisService("localhost:6379", "", 0)
	musicHandler := handlers.NewMusicSearchHandler(externalAPIService, redisService, cacheService)

	// Initialize enhanced playlist service and handler
	playlistEnhancedService := services.NewPlaylistEnhancedService(integratedMongoClient.Database("gruvit"), userService)
	playlistEnhancedHandler := handlers.NewPlaylistEnhancedHandler(playlistEnhancedService, userService)

	// Initialize social service and handler
	socialService := services.NewSocialService(integratedMongoClient.Database("gruvit"))
	socialHandler := handlers.NewSocialHandler(socialService)

	// Initialize recommendation service and handler
	_ = services.NewRecommendationService(integratedMongoClient.Database("gruvit"), userService)

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check auth service health
		authHealthy := true
		if err := integratedAuthClient.HealthCheck(); err != nil {
			authHealthy = false
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"service":      "music-api-integrated",
			"version":      "1.0.0",
			"auth_service": authHealthy,
		})
	})

	// Auth endpoints - delegate to Java service
	r.POST("/auth/login", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["auth"]), authHandler.Login)
	r.POST("/auth/register", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["auth"]), authHandler.Register)
	r.POST("/auth/refresh", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["auth"]), authHandler.RefreshToken)
	r.POST("/auth/validate", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["auth"]), authHandler.ValidateToken)
	r.GET("/auth/profile", middleware.AuthMiddleware(integratedAuthClient), authHandler.GetUserProfile)
	r.PUT("/auth/profile", middleware.AuthMiddleware(integratedAuthClient), authHandler.UpdateUserProfile)
	r.POST("/auth/logout", middleware.AuthMiddleware(integratedAuthClient), authHandler.Logout)

	// WebSocket endpoints
	r.GET("/ws", middleware.AuthMiddleware(integratedAuthClient), wsHandler.HandleWebSocket)
	r.GET("/ws/public", wsHandler.HandlePublicWebSocket)
	r.GET("/api/ws/stats", middleware.AuthMiddleware(integratedAuthClient), wsHandler.GetWebSocketStats)

	// Advanced search endpoint with filters, sorting, and pagination
	r.GET("/search", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["search"]), func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		// Parse pagination parameters
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		// Parse filter parameters
		genre := c.Query("genre")
		artist := c.Query("artist")
		source := c.Query("source") // jamendo, musicbrainz, or all
		minDuration := c.Query("min_duration")
		maxDuration := c.Query("max_duration")

		// Parse sort parameters
		sortBy := c.DefaultQuery("sort_by", "relevance")  // relevance, title, artist, duration, date
		sortOrder := c.DefaultQuery("sort_order", "desc") // asc, desc

		var allTracks []models.Track

		// Check cache first
		cacheKey := fmt.Sprintf("search:%s:%d:%d:%s:%s:%s:%s:%s:%s", query, page, limit, genre, artist, source, minDuration, maxDuration, sortBy)
		if integratedRedisClient != nil {
			if cached, err := integratedRedisClient.Get(integratedRedisCtx, cacheKey).Result(); err == nil {
				json.Unmarshal([]byte(cached), &allTracks)
				c.JSON(http.StatusOK, gin.H{
					"query":   query,
					"results": allTracks,
					"total":   len(allTracks),
					"page":    page,
					"limit":   limit,
					"offset":  offset,
					"filters": gin.H{"genre": genre, "artist": artist, "source": source, "min_duration": minDuration, "max_duration": maxDuration},
					"sort":    gin.H{"by": sortBy, "order": sortOrder},
				})
				return
			}
		}

		// Search Jamendo
		if source == "" || source == "jamendo" || source == "all" {
			jamURL := "https://api.jamendo.com/v3.0/tracks/?client_id=" + os.Getenv("JAMENDO_API_KEY") + "&format=json&search=" + query + "&limit=" + strconv.Itoa(limit*2) // Get more to filter
			resp, err := http.Get(jamURL)
			if err == nil && resp.StatusCode == 200 {
				var jamData struct {
					Results []struct {
						ID         string `json:"id"`
						Name       string `json:"name"`
						ArtistName string `json:"artist_name"`
						AlbumName  string `json:"album_name"`
						Duration   int    `json:"duration"`
						Audio      string `json:"audio"`
						Image      string `json:"album_image"`
						Genre      string `json:"musicinfo_genres"`
					} `json:"results"`
				}
				json.NewDecoder(resp.Body).Decode(&jamData)
				resp.Body.Close()

				for _, t := range jamData.Results {
					// Apply filters
					if genre != "" && !strings.Contains(strings.ToLower(t.Genre), strings.ToLower(genre)) {
						continue
					}
					if artist != "" && !strings.Contains(strings.ToLower(t.ArtistName), strings.ToLower(artist)) {
						continue
					}
					if minDuration != "" {
						if minDur, err := strconv.Atoi(minDuration); err == nil && t.Duration < minDur {
							continue
						}
					}
					if maxDuration != "" {
						if maxDur, err := strconv.Atoi(maxDuration); err == nil && t.Duration > maxDur {
							continue
						}
					}

					// Don't generate stream URL here - let the streaming service handle it
					// This prevents URL validation issues and ensures consistent streaming
					allTracks = append(allTracks, models.Track{
						ID:        t.ID,
						Title:     t.Name,
						Artist:    t.ArtistName,
						Album:     t.AlbumName,
						Duration:  t.Duration,
						StreamURL: "", // Will be generated by streaming service when needed
						ImageURL:  t.Image,
						Genre:     t.Genre,
						Source:    "jamendo",
						UpdatedAt: time.Now(),
					})
				}
			}
		}

		// Search MusicBrainz
		if source == "" || source == "musicbrainz" || source == "all" {
			mbURL := "https://musicbrainz.org/ws/2/recording?query=recording:" + query + "&fmt=json&limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
			req, _ := http.NewRequest("GET", mbURL, nil)
			req.Header.Set("User-Agent", os.Getenv("USER_AGENT"))
			req.Header.Set("Accept", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err == nil && resp.StatusCode == 200 {
				var mbData struct {
					Recordings []struct {
						ID           string `json:"id"`
						Title        string `json:"title"`
						Length       int    `json:"length"`
						ArtistCredit []struct {
							Name string `json:"name"`
						} `json:"artist-credit"`
						Releases []struct {
							Title string `json:"title"`
						} `json:"releases"`
					} `json:"recordings"`
				}
				json.NewDecoder(resp.Body).Decode(&mbData)
				resp.Body.Close()

				for _, rec := range mbData.Recordings {
					artist := ""
					if len(rec.ArtistCredit) > 0 {
						artist = rec.ArtistCredit[0].Name
					}
					album := ""
					if len(rec.Releases) > 0 {
						album = rec.Releases[0].Title
					}

					// Apply filters
					if artist != "" && !strings.Contains(strings.ToLower(artist), strings.ToLower(artist)) {
						continue
					}
					if minDuration != "" {
						if minDur, err := strconv.Atoi(minDuration); err == nil && rec.Length < minDur {
							continue
						}
					}
					if maxDuration != "" {
						if maxDur, err := strconv.Atoi(maxDuration); err == nil && rec.Length > maxDur {
							continue
						}
					}

					allTracks = append(allTracks, models.Track{
						ID:        rec.ID,
						Title:     rec.Title,
						Artist:    artist,
						Album:     album,
						Duration:  rec.Length,
						Source:    "musicbrainz",
						UpdatedAt: time.Now(),
					})
				}
			}
		}

		// Apply sorting
		switch sortBy {
		case "title":
			if sortOrder == "asc" {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Title < allTracks[j].Title })
			} else {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Title > allTracks[j].Title })
			}
		case "artist":
			if sortOrder == "asc" {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Artist < allTracks[j].Artist })
			} else {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Artist > allTracks[j].Artist })
			}
		case "duration":
			if sortOrder == "asc" {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Duration < allTracks[j].Duration })
			} else {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].Duration > allTracks[j].Duration })
			}
		case "date":
			if sortOrder == "asc" {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].UpdatedAt.Before(allTracks[j].UpdatedAt) })
			} else {
				sort.Slice(allTracks, func(i, j int) bool { return allTracks[i].UpdatedAt.After(allTracks[j].UpdatedAt) })
			}
		}

		// Apply pagination
		total := len(allTracks)
		start := offset
		end := offset + limit
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		if start < 0 {
			start = 0
		}
		paginatedTracks := allTracks[start:end]

		// Cache results
		if integratedRedisClient != nil && len(paginatedTracks) > 0 {
			data, _ := json.Marshal(paginatedTracks)
			integratedRedisClient.Set(integratedRedisCtx, cacheKey, data, time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   query,
			"results": paginatedTracks,
			"total":   total,
			"page":    page,
			"limit":   limit,
			"offset":  offset,
			"filters": gin.H{"genre": genre, "artist": artist, "source": source, "min_duration": minDuration, "max_duration": maxDuration},
			"sort":    gin.H{"by": sortBy, "order": sortOrder},
		})
	})

	// Enhanced music search endpoints
	r.GET("/music/search", func(c *gin.Context) {
		// Same as /search but with additional features
		c.Redirect(http.StatusMovedPermanently, "/search?"+c.Request.URL.RawQuery)
	})

	// Music search by artist
	r.GET("/music/artist", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["search"]), func(c *gin.Context) {
		artist := c.Query("artist")
		if artist == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Artist parameter is required"})
			return
		}

		limitStr := c.DefaultQuery("limit", "20")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 20
		}

		// Search for tracks by artist
		var allTracks []models.Track

		// Check cache first
		cacheKey := "artist:" + artist + ":" + limitStr
		if integratedRedisClient != nil {
			if cached, err := integratedRedisClient.Get(integratedRedisCtx, cacheKey).Result(); err == nil {
				json.Unmarshal([]byte(cached), &allTracks)
				c.JSON(http.StatusOK, gin.H{
					"artist":  artist,
					"results": allTracks,
					"total":   len(allTracks),
				})
				return
			}
		}

		// Search Jamendo for artist tracks
		jamURL := "https://api.jamendo.com/v3.0/tracks/?client_id=" + os.Getenv("JAMENDO_API_KEY") + "&format=json&artist_name=" + artist + "&limit=" + strconv.Itoa(limit)
		resp, err := http.Get(jamURL)
		if err == nil && resp.StatusCode == 200 {
			var jamData struct {
				Results []struct {
					ID         string `json:"id"`
					Name       string `json:"name"`
					ArtistName string `json:"artist_name"`
					AlbumName  string `json:"album_name"`
					Duration   int    `json:"duration"`
					Audio      string `json:"audio"`
					Image      string `json:"album_image"`
					Genre      string `json:"musicinfo_genres"`
				} `json:"results"`
			}
			json.NewDecoder(resp.Body).Decode(&jamData)
			resp.Body.Close()

			for _, t := range jamData.Results {
				// Don't generate stream URL here - let the streaming service handle it
				allTracks = append(allTracks, models.Track{
					ID:        t.ID,
					Title:     t.Name,
					Artist:    t.ArtistName,
					Album:     t.AlbumName,
					Duration:  t.Duration,
					StreamURL: "", // Will be generated by streaming service when needed
					ImageURL:  t.Image,
					Genre:     t.Genre,
					Source:    "jamendo",
				})
			}
		}

		// Cache results
		if integratedRedisClient != nil && len(allTracks) > 0 {
			data, _ := json.Marshal(allTracks)
			integratedRedisClient.Set(integratedRedisCtx, cacheKey, data, time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"artist":  artist,
			"results": allTracks,
			"total":   len(allTracks),
		})
	})

	// Music search by genre
	r.GET("/music/genre", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["search"]), func(c *gin.Context) {
		genre := c.Query("genre")
		if genre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Genre parameter is required"})
			return
		}

		limitStr := c.DefaultQuery("limit", "20")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 20
		}

		// Search for tracks by genre
		var allTracks []models.Track

		// Check cache first
		cacheKey := "genre:" + genre + ":" + limitStr
		if integratedRedisClient != nil {
			if cached, err := integratedRedisClient.Get(integratedRedisCtx, cacheKey).Result(); err == nil {
				json.Unmarshal([]byte(cached), &allTracks)
				c.JSON(http.StatusOK, gin.H{
					"genre":   genre,
					"results": allTracks,
					"total":   len(allTracks),
				})
				return
			}
		}

		// Search Jamendo for genre tracks
		jamURL := "https://api.jamendo.com/v3.0/tracks/?client_id=" + os.Getenv("JAMENDO_API_KEY") + "&format=json&tags=" + genre + "&limit=" + strconv.Itoa(limit)
		resp, err := http.Get(jamURL)
		if err == nil && resp.StatusCode == 200 {
			var jamData struct {
				Results []struct {
					ID         string `json:"id"`
					Name       string `json:"name"`
					ArtistName string `json:"artist_name"`
					AlbumName  string `json:"album_name"`
					Duration   int    `json:"duration"`
					Audio      string `json:"audio"`
					Image      string `json:"album_image"`
					Genre      string `json:"musicinfo_genres"`
				} `json:"results"`
			}
			json.NewDecoder(resp.Body).Decode(&jamData)
			resp.Body.Close()

			for _, t := range jamData.Results {
				// Don't generate stream URL here - let the streaming service handle it
				allTracks = append(allTracks, models.Track{
					ID:        t.ID,
					Title:     t.Name,
					Artist:    t.ArtistName,
					Album:     t.AlbumName,
					Duration:  t.Duration,
					StreamURL: "", // Will be generated by streaming service when needed
					ImageURL:  t.Image,
					Genre:     t.Genre,
					Source:    "jamendo",
				})
			}
		}

		// Cache results
		if integratedRedisClient != nil && len(allTracks) > 0 {
			data, _ := json.Marshal(allTracks)
			integratedRedisClient.Set(integratedRedisCtx, cacheKey, data, time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"genre":   genre,
			"results": allTracks,
			"total":   len(allTracks),
		})
	})

	// Popular tracks endpoint
	r.GET("/music/popular", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["search"]), func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "20")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 20
		}

		// Get popular tracks from Jamendo
		var allTracks []models.Track

		// Check cache first
		cacheKey := "popular:" + limitStr
		if integratedRedisClient != nil {
			if cached, err := integratedRedisClient.Get(integratedRedisCtx, cacheKey).Result(); err == nil {
				json.Unmarshal([]byte(cached), &allTracks)
				c.JSON(http.StatusOK, gin.H{
					"results": allTracks,
					"total":   len(allTracks),
				})
				return
			}
		}

		// Get popular tracks from Jamendo
		jamURL := "https://api.jamendo.com/v3.0/tracks/?client_id=" + os.Getenv("JAMENDO_API_KEY") + "&format=json&order=popularity_total&limit=" + strconv.Itoa(limit)
		resp, err := http.Get(jamURL)
		if err == nil && resp.StatusCode == 200 {
			var jamData struct {
				Results []struct {
					ID         string `json:"id"`
					Name       string `json:"name"`
					ArtistName string `json:"artist_name"`
					AlbumName  string `json:"album_name"`
					Duration   int    `json:"duration"`
					Audio      string `json:"audio"`
					Image      string `json:"album_image"`
					Genre      string `json:"musicinfo_genres"`
				} `json:"results"`
			}
			json.NewDecoder(resp.Body).Decode(&jamData)
			resp.Body.Close()

			for _, t := range jamData.Results {
				// Don't generate stream URL here - let the streaming service handle it
				allTracks = append(allTracks, models.Track{
					ID:        t.ID,
					Title:     t.Name,
					Artist:    t.ArtistName,
					Album:     t.AlbumName,
					Duration:  t.Duration,
					StreamURL: "", // Will be generated by streaming service when needed
					ImageURL:  t.Image,
					Genre:     t.Genre,
					Source:    "jamendo",
				})
			}
		}

		// Cache results
		if integratedRedisClient != nil && len(allTracks) > 0 {
			data, _ := json.Marshal(allTracks)
			integratedRedisClient.Set(integratedRedisCtx, cacheKey, data, time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"results": allTracks,
			"total":   len(allTracks),
		})
	})

	// Track details endpoint
	r.GET("/music/track/:id", func(c *gin.Context) {
		trackID := c.Param("id")
		source := c.Query("source")

		if trackID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Track ID is required"})
			return
		}

		// For now, return a basic track structure
		// In a real implementation, you would fetch from database or external API
		track := models.Track{
			ID:        trackID,
			Title:     "Unknown Track",
			Artist:    "Unknown Artist",
			Source:    source,
			UpdatedAt: time.Now(),
		}

		c.JSON(http.StatusOK, track)
	})

	// Public stream endpoint for Jamendo tracks (no auth required)
	r.GET("/stream/:trackId", rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["stream"]), func(c *gin.Context) {
		trackId := c.Param("trackId")
		source := c.DefaultQuery("source", "jamendo") // Default to jamendo for public access

		// Only allow Jamendo for public access
		if source != "jamendo" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only Jamendo tracks are available for public streaming"})
			return
		}

		// Use the streaming service
		response, err := streamingService.GetStreamURL(trackId, source)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return in the format expected by frontend: { stream_url: string }
		c.JSON(http.StatusOK, gin.H{
			"stream_url": response.StreamURL,
		})
	})

	// Authenticated stream endpoint - requires authentication
	r.GET("/api/stream/:trackId", middleware.AuthMiddleware(integratedAuthClient), streamHandler.GetStreamURL)

	// Playlist endpoints - require authentication
	r.GET("/api/playlists", middleware.AuthMiddleware(integratedAuthClient), rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["playlist"]), func(c *gin.Context) {
		username, _ := c.Get("username")
		coll := integratedMongoClient.Database("gruvit").Collection("playlists")

		cursor, err := coll.Find(context.Background(), bson.M{
			"$or": []bson.M{
				{"owner": username},
				{"is_public": true},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists"})
			return
		}
		defer cursor.Close(context.Background())

		var playlists []models.Playlist
		cursor.All(context.Background(), &playlists)

		c.JSON(http.StatusOK, gin.H{"playlists": playlists})
	})

	r.POST("/api/playlists", middleware.AuthMiddleware(integratedAuthClient), rateLimitMiddleware.RateLimit(rateLimitMiddleware.DefaultConfigs()["playlist"]), func(c *gin.Context) {
		username, _ := c.Get("username")
		var playlist models.Playlist
		if err := c.BindJSON(&playlist); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		playlist.Owner = username.(string)
		playlist.CreatedAt = time.Now()
		playlist.UpdatedAt = time.Now()

		coll := integratedMongoClient.Database("gruvit").Collection("playlists")
		result, err := coll.InsertOne(context.Background(), playlist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
			return
		}

		playlist.ID = result.InsertedID.(string)
		c.JSON(http.StatusCreated, playlist)
	})

	// Enhanced Playlist Features - Collaborative Playlists
	r.POST("/api/playlists/collaborative", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.CreateCollaborativePlaylist)
	r.POST("/api/playlists/:playlistId/collaborators", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.AddCollaborator)
	r.DELETE("/api/playlists/:playlistId/collaborators/:userId", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.RemoveCollaborator)
	r.GET("/api/playlists/:playlistId/collaborators", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.GetCollaborators)

	// Enhanced Playlist Features - Track Management
	r.POST("/api/playlists/:playlistId/tracks", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.AddTrackToPlaylist)
	r.DELETE("/api/playlists/:playlistId/tracks/:trackId", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.RemoveTrackFromPlaylist)

	// Enhanced Playlist Features - Following and Liking
	r.POST("/api/playlists/:playlistId/follow", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.FollowPlaylist)
	r.DELETE("/api/playlists/:playlistId/follow", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.UnfollowPlaylist)
	r.POST("/api/playlists/:playlistId/like", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.LikePlaylist)
	r.DELETE("/api/playlists/:playlistId/like", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.UnlikePlaylist)

	// Enhanced Playlist Features - Sharing
	r.POST("/api/playlists/:playlistId/share", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.SharePlaylist)
	r.GET("/api/playlists/shared/:token", playlistEnhancedHandler.GetPlaylistByShareToken)

	// Enhanced Playlist Features - Recommendations and Discovery
	r.GET("/api/playlists/recommendations", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.GetPlaylistRecommendations)
	r.GET("/api/playlists/followed", middleware.AuthMiddleware(integratedAuthClient), playlistEnhancedHandler.GetFollowedPlaylists)
	r.GET("/api/playlists/public", playlistEnhancedHandler.GetPublicPlaylists)

	// Social Features - Social Feed and Activity
	r.GET("/api/social/feed", middleware.AuthMiddleware(integratedAuthClient), socialHandler.GetUserFeed)
	r.GET("/api/social/activities/:userId", socialHandler.GetUserActivities)
	r.POST("/api/social/follow/:userId", middleware.AuthMiddleware(integratedAuthClient), socialHandler.FollowUser)
	r.DELETE("/api/social/follow/:userId", middleware.AuthMiddleware(integratedAuthClient), socialHandler.UnfollowUser)
	r.GET("/api/social/stats/:userId", socialHandler.GetUserStats)
	r.GET("/api/social/notifications", middleware.AuthMiddleware(integratedAuthClient), socialHandler.GetNotifications)
	r.PUT("/api/social/notifications/:notificationId/read", middleware.AuthMiddleware(integratedAuthClient), socialHandler.MarkNotificationAsRead)
	r.POST("/api/social/activity", middleware.AuthMiddleware(integratedAuthClient), socialHandler.RecordActivity)

	// User profile endpoint - requires authentication
	r.GET("/api/profile", middleware.AuthMiddleware(integratedAuthClient), func(c *gin.Context) {
		username, _ := c.Get("username")
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		// Get user profile from database
		profile, err := userService.GetUserProfile(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
			return
		}

		// Get user stats
		stats, err := userService.GetUserStats(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
			return
		}

		// Create user object in the format expected by frontend
		userData := gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
		}

		if profile != nil {
			userData["display_name"] = profile.DisplayName
			userData["bio"] = profile.Bio
			userData["avatar"] = profile.Avatar
			userData["location"] = profile.Location
			userData["website"] = profile.Website
			userData["created_at"] = profile.CreatedAt
			userData["updated_at"] = profile.UpdatedAt
		}

		if stats != nil {
			userData["total_plays"] = stats.TotalPlays
			userData["total_playlists"] = stats.TotalPlaylists
			userData["total_favorites"] = stats.TotalFavorites
			userData["total_following"] = stats.TotalFollowing
			userData["total_followers"] = stats.TotalFollowers
			userData["last_active"] = stats.LastActive
		}

		// Return in the format expected by frontend: { user: User }
		c.JSON(http.StatusOK, gin.H{
			"user": userData,
		})
	})

	// User endpoints - require authentication
	r.GET("/api/user/favorites", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserFavorites)
	r.GET("/api/user/history", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserListeningHistory)
	r.GET("/api/user/top-artists", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserTopArtists)
	r.GET("/api/user/top-tracks", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserTopTracks)
	r.GET("/api/user/followings", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserFollowings)
	r.GET("/api/user/followers", middleware.AuthMiddleware(integratedAuthClient), userHandler.GetUserFollowers)
	r.POST("/api/user/follow/artist/:artistId", middleware.AuthMiddleware(integratedAuthClient), userHandler.FollowArtist)
	r.DELETE("/api/user/follow/artist/:artistId", middleware.AuthMiddleware(integratedAuthClient), userHandler.UnfollowArtist)
	r.POST("/api/user/favorites/:trackId", middleware.AuthMiddleware(integratedAuthClient), userHandler.AddToFavorites)
	r.DELETE("/api/user/favorites/:trackId", middleware.AuthMiddleware(integratedAuthClient), userHandler.RemoveFromFavorites)
	r.POST("/api/user/record-play", middleware.AuthMiddleware(integratedAuthClient), userHandler.RecordPlay)

	// User analytics endpoints - require authentication
	r.GET("/api/user/stats", middleware.AuthMiddleware(integratedAuthClient), func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		// Get user stats
		stats, err := userService.GetUserStats(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
			return
		}

		// Get top artists
		topArtists, err := userService.GetUserTopArtists(userID.(string), 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top artists"})
			return
		}

		// Get top tracks
		topTracks, err := userService.GetUserTopTracks(userID.(string), 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top tracks"})
			return
		}

		// Create response in the format expected by frontend
		response := gin.H{
			"total_plays":    stats.TotalPlays,
			"unique_artists": len(topArtists),
			"unique_tracks":  len(topTracks),
			"top_artists":    topArtists,
			"top_tracks":     topTracks,
		}

		c.JSON(http.StatusOK, response)
	})

	// Advanced music features endpoints
	r.GET("/api/music/albums/:albumId", musicHandler.GetAlbumDetails)
	r.GET("/api/music/artists/:artistId", musicHandler.GetArtistDetails)
	r.GET("/api/music/genres", musicHandler.GetGenres)
	r.GET("/api/music/genres/:genre/tracks", musicHandler.GetTracksByGenre)
	r.GET("/api/music/search/advanced", musicHandler.AdvancedSearch)
	r.GET("/api/music/trending", musicHandler.GetTrendingTracks)
	r.GET("/api/music/popular", musicHandler.GetPopularTracks)
	r.GET("/api/music/discover", musicHandler.GetDiscoverTracks)
	r.GET("/api/music/artists/:artistId/similar", musicHandler.GetSimilarArtists)
	r.GET("/api/music/tracks/:trackId/similar", musicHandler.GetSimilarTracks)

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting integrated music API server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}

	log.Println("Server exited")
}
