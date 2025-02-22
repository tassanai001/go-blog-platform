package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
	"go-blog-platform/config"
	"go-blog-platform/internal/handlers"
	"go-blog-platform/internal/middleware"
	"go-blog-platform/internal/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set up MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get database instance
	db := client.Database(cfg.MongoDB.Database)

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Initialize services
	emailService := services.NewEmailService()
	mediaService := services.NewMediaService(uploadsDir, cfg.BaseURL)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db, cfg.JWT.Secret, emailService, mediaService, cfg.BaseURL)
	postHandler := handlers.NewPostHandler(db, mediaService)

	// Initialize router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	r.Use(static.Serve("/", static.LocalFile("ui/build", true)))

	// API routes
	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/password-reset/request", userHandler.RequestPasswordReset)
			auth.POST("/password-reset/reset", userHandler.ResetPassword)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("", userHandler.ListUsers)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.PUT("/:id/role", userHandler.UpdateUserRole)
				users.DELETE("/:id", userHandler.DeleteUser)
			}

			// Post routes
			posts := protected.Group("/posts")
			{
				posts.GET("", postHandler.List)
				posts.POST("", postHandler.Create)
				posts.GET("/:id", postHandler.Get)
				posts.PUT("/:id", postHandler.Update)
				posts.DELETE("/:id", postHandler.Delete)
			}

			// Media routes
			media := protected.Group("/media")
			{
				media.POST("", func(c *gin.Context) {
					file, err := c.FormFile("file")
					if err != nil {
						c.JSON(400, gin.H{"error": "Failed to get file"})
						return
					}

					userID, _ := c.Get("user_id")
					filePath, thumbnails, err := mediaService.SaveFile(file, userID.(string))
					if err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}

					c.JSON(200, gin.H{
						"path": filePath,
						"thumbnails": thumbnails,
					})
				})
				media.DELETE("/:path", func(c *gin.Context) {
					path := c.Param("path")
					if err := mediaService.DeleteFile(path); err != nil {
						c.JSON(500, gin.H{"error": err.Error()})
						return
					}
					c.JSON(200, gin.H{"message": "File deleted successfully"})
				})
			}
		}
	}

	// Serve media files
	r.Static("/media", uploadsDir)

	// Start server
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
