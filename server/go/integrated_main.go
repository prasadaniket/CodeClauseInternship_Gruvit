package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

func mainIntegrated() {
	// Load .env
	if err := godotenv.Load("config.dev.env"); err != nil {
		log.Println("No config.dev.env file found")
	}

	// Initialize auth client
	authClient := services.NewAuthClient()

	// Check auth service health
	if err := authClient.HealthCheck(); err != nil {
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

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Test MongoDB connection
	err = mongoClient.Ping(context.Background(), nil)
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Continuing without caching", err)
		redisClient = nil
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

	// Initialize handlers
	authHandler := handlers.NewIntegratedAuthHandler(authClient)

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check auth service health
		authHealthy := true
		if err := authClient.HealthCheck(); err != nil {
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
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/refresh", authHandler.RefreshToken)
	r.POST("/auth/validate", authHandler.ValidateToken)
	r.GET("/auth/profile", middleware.AuthMiddleware(authClient), authHandler.GetUserProfile)
	r.PUT("/auth/profile", middleware.AuthMiddleware(authClient), authHandler.UpdateUserProfile)
	r.POST("/auth/logout", middleware.AuthMiddleware(authClient), authHandler.Logout)

	// Search endpoint with pagination (MusicBrainz + Jamendo)
	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		// Parse pagination parameters
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 20
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		var allTracks []models.Track

		// Check cache first
		cacheKey := "search:" + query + ":" + limitStr + ":" + offsetStr
		if redisClient != nil {
			if cached, err := redisClient.Get(context.Background(), cacheKey).Result(); err == nil {
				json.Unmarshal([]byte(cached), &allTracks)
				c.JSON(http.StatusOK, gin.H{
					"query":   query,
					"results": allTracks,
					"total":   len(allTracks),
					"page":    (offset / limit) + 1,
					"limit":   limit,
					"offset":  offset,
				})
				return
			}
		}

		// MusicBrainz query
		mbLimit := limit / 2 // Split between APIs
		if mbLimit > 0 {
			mbURL := "https://musicbrainz.org/ws/2/recording?query=recording:" + query + "&fmt=json&limit=" + strconv.Itoa(mbLimit) + "&offset=" + strconv.Itoa(offset/2)
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
					allTracks = append(allTracks, models.Track{
						ID:       rec.ID,
						Title:    rec.Title,
						Artist:   artist,
						Album:    album,
						Duration: rec.Length,
						Source:   "musicbrainz",
					})
				}
			}
		}

		// Jamendo query
		jamLimit := limit - len(allTracks)
		if jamLimit > 0 {
			jamURL := "https://api.jamendo.com/v3.0/tracks/?client_id=" + os.Getenv("JAMENDO_API_KEY") + "&format=json&search=" + query + "&limit=" + strconv.Itoa(jamLimit) + "&offset=" + strconv.Itoa(offset/2)
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
					// Don't use t.Audio directly - let the streaming service handle URL generation
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
		}

		// Cache results
		if redisClient != nil && len(allTracks) > 0 {
			data, _ := json.Marshal(allTracks)
			redisClient.Set(context.Background(), cacheKey, data, time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   query,
			"results": allTracks,
			"total":   len(allTracks),
			"page":    (offset / limit) + 1,
			"limit":   limit,
			"offset":  offset,
		})
	})

	// Enhanced music search endpoints
	r.GET("/music/search", func(c *gin.Context) {
		// Same as /search but with additional features
		c.Redirect(http.StatusMovedPermanently, "/search?"+c.Request.URL.RawQuery)
	})

	// Public stream endpoint placeholder (no auth required)
	r.GET("/stream/:trackId", func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Streaming service not available in this build"})
	})

	// Authenticated stream endpoint placeholder
	r.GET("/api/stream/:trackId", middleware.AuthMiddleware(authClient), func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Streaming service not available in this build"})
	})

	// Playlist endpoints - require authentication
	r.GET("/api/playlists", middleware.AuthMiddleware(authClient), func(c *gin.Context) {
		username, _ := c.Get("username")
		coll := mongoClient.Database("gruvit").Collection("playlists")

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

	r.POST("/api/playlists", middleware.AuthMiddleware(authClient), func(c *gin.Context) {
		username, _ := c.Get("username")
		var playlist models.Playlist
		if err := c.BindJSON(&playlist); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		playlist.Owner = username.(string)
		playlist.CreatedAt = time.Now()
		playlist.UpdatedAt = time.Now()

		coll := mongoClient.Database("gruvit").Collection("playlists")
		result, err := coll.InsertOne(context.Background(), playlist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
			return
		}

		playlist.ID = result.InsertedID.(string)
		c.JSON(http.StatusCreated, playlist)
	})

	// User profile endpoint - requires authentication
	r.GET("/api/profile", middleware.AuthMiddleware(authClient), func(c *gin.Context) {
		username, _ := c.Get("username")
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
			"joined":   time.Now().Add(-30 * 24 * time.Hour), // Mock data
		})
	})

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
