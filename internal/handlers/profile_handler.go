package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"

    "go-blog-platform/internal/models"
)

type ProfileHandler struct {
    collection *mongo.Collection
}

func NewProfileHandler(db *mongo.Database) *ProfileHandler {
    return &ProfileHandler{
        collection: db.Collection("profiles"),
    }
}

// GetProfile retrieves a user's profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
    userID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    ctx := context.Background()
    var profile models.Profile
    err = h.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&profile)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
        return
    }

    c.JSON(http.StatusOK, profile)
}

// GetMyProfile retrieves the current user's profile
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    objID, err := primitive.ObjectIDFromHex(userID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    ctx := context.Background()
    var profile models.Profile
    err = h.collection.FindOne(ctx, bson.M{"user_id": objID}).Decode(&profile)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            // If profile doesn't exist, create an empty one
            profile = models.Profile{
                ID:        primitive.NewObjectID(),
                UserID:    objID,
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            }
            _, err = h.collection.InsertOne(ctx, profile)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
                return
            }
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
            return
        }
    }

    c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the current user's profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    objID, err := primitive.ObjectIDFromHex(userID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var profile models.Profile
    if err := c.ShouldBindJSON(&profile); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Ensure we're updating the correct profile
    profile.UserID = objID
    profile.UpdatedAt = time.Now()

    ctx := context.Background()
    result, err := h.collection.UpdateOne(
        ctx,
        bson.M{"user_id": objID},
        bson.M{
            "$set": bson.M{
                "full_name":    profile.FullName,
                "bio":         profile.Bio,
                "avatar":      profile.Avatar,
                "website":     profile.Website,
                "location":    profile.Location,
                "social_links": profile.SocialLinks,
                "updated_at":   profile.UpdatedAt,
            },
        },
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
        return
    }

    if result.MatchedCount == 0 {
        // If profile doesn't exist, create a new one
        profile.ID = primitive.NewObjectID()
        profile.CreatedAt = time.Now()
        _, err = h.collection.InsertOne(ctx, profile)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
            return
        }
    }

    c.JSON(http.StatusOK, profile)
}

// DeleteProfile deletes the current user's profile
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    objID, err := primitive.ObjectIDFromHex(userID.(string))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    ctx := context.Background()
    result, err := h.collection.DeleteOne(ctx, bson.M{"user_id": objID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
        return
    }

    c.Status(http.StatusNoContent)
}
