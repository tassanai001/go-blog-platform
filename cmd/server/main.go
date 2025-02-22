package main

import (
    "context"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "go-blog-platform/config"
    "go-blog-platform/internal/constants"
    "go-blog-platform/internal/handlers"
    "go-blog-platform/internal/middleware"
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

    // Initialize router
    r := gin.Default()

    // Initialize handlers
    postHandler := handlers.NewPostHandler(db)
    userHandler := handlers.NewUserHandler(db, cfg.JWT.Secret)
    profileHandler := handlers.NewProfileHandler(db)

    // Public routes
    r.POST("/api/register", userHandler.Register)
    r.POST("/api/login", userHandler.Login)
    r.GET("/api/profiles/:id", profileHandler.GetProfile) // Public profile view

    // Protected routes
    api := r.Group("/api")
    api.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
    {
        // Profile routes
        profile := api.Group("/profile")
        {
            profile.GET("/me", profileHandler.GetMyProfile)
            profile.PUT("/me", profileHandler.UpdateProfile)
            profile.DELETE("/me", profileHandler.DeleteProfile)
        }

        // Reader routes (accessible by all authenticated users)
        posts := api.Group("/posts")
        {
            posts.GET("/", postHandler.List)                            // All users can list posts
            posts.GET("/:id", postHandler.Get)                         // All users can view posts
            posts.POST("/", middleware.RequireRole(constants.RoleAuthor), postHandler.Create)    // Only authors and admins can create
            posts.PUT("/:id", middleware.IsAuthorOrAdmin(), postHandler.Update)                 // Only authors and admins can update
            posts.DELETE("/:id", middleware.IsAdmin(), postHandler.Delete)                      // Only admins can delete
        }

        // Author routes
        author := api.Group("/author")
        author.Use(middleware.RequireRole(constants.RoleAuthor))
        {
            author.GET("/drafts", postHandler.ListDrafts)              // Authors can view their drafts
            author.POST("/drafts", postHandler.CreateDraft)            // Authors can create drafts
            // Add more author-specific routes here
        }

        // Admin routes
        admin := api.Group("/admin")
        admin.Use(middleware.IsAdmin())
        {
            admin.GET("/users", userHandler.ListUsers)                 // Admins can list all users
            admin.PUT("/users/:id/role", userHandler.UpdateUserRole)   // Admins can update user roles
            admin.DELETE("/users/:id", userHandler.DeleteUser)         // Admins can delete users
            // Add more admin-specific routes here
        }
    }

    // Start server
    r.Run(":" + cfg.Server.Port)
}
