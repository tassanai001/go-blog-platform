package handlers

import (
    "context"
    "net/http"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "go-blog-platform/internal/models"
    "go-blog-platform/internal/services"
)

type MediaHandler struct {
    collection   *mongo.Collection
    mediaService *services.MediaService
    baseURL      string
}

func NewMediaHandler(db *mongo.Database, mediaService *services.MediaService, baseURL string) *MediaHandler {
    return &MediaHandler{
        collection:   db.Collection("media"),
        mediaService: mediaService,
        baseURL:      baseURL,
    }
}

// UploadMedia handles file uploads
func (h *MediaHandler) UploadMedia(c *gin.Context) {
    // Get user ID from context
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Parse multipart form
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()

    // Validate file
    if err := h.mediaService.ValidateFile(header); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Save file and create thumbnails
    filePath, thumbnails, err := h.mediaService.SaveFile(header, userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    // Create media record
    objID, _ := primitive.ObjectIDFromHex(userID.(string))
    media := models.Media{
        ID:        primitive.NewObjectID(),
        UserID:    objID,
        FileName:  header.Filename,
        FileType:  filepath.Ext(header.Filename),
        MimeType:  header.Header.Get("Content-Type"),
        Size:      header.Size,
        Path:      filePath,
        URL:       h.generateURL(filePath),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Add thumbnails if any
    for _, thumb := range thumbnails {
        media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
            Size: thumb.Size,
            Path: thumb.Path,
            URL:  h.generateURL(thumb.Path),
        })
    }

    // Save to database
    ctx := context.Background()
    _, err = h.collection.InsertOne(ctx, media)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save media record"})
        return
    }

    c.JSON(http.StatusCreated, media)
}

// ListMedia returns a list of media files
func (h *MediaHandler) ListMedia(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    objID, _ := primitive.ObjectIDFromHex(userID.(string))
    
    ctx := context.Background()
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
    
    cursor, err := h.collection.Find(ctx, bson.M{"user_id": objID}, opts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch media"})
        return
    }
    defer cursor.Close(ctx)

    var media []models.Media
    if err := cursor.All(ctx, &media); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode media"})
        return
    }

    c.JSON(http.StatusOK, media)
}

// GetMedia returns a single media file
func (h *MediaHandler) GetMedia(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
        return
    }

    ctx := context.Background()
    var media models.Media
    err = h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&media)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch media"})
        return
    }

    c.JSON(http.StatusOK, media)
}

// DeleteMedia deletes a media file
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
        return
    }

    objID, _ := primitive.ObjectIDFromHex(userID.(string))
    
    ctx := context.Background()
    var media models.Media
    err = h.collection.FindOne(ctx, bson.M{
        "_id": id,
        "user_id": objID,
    }).Decode(&media)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch media"})
        return
    }

    // Delete file and thumbnails
    if err := h.mediaService.DeleteFile(media.Path); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
        return
    }

    // Delete from database
    _, err = h.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media record"})
        return
    }

    c.Status(http.StatusNoContent)
}

// UpdateMedia updates media metadata
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media ID"})
        return
    }

    var metadata models.MediaMetadata
    if err := c.ShouldBindJSON(&metadata); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    objID, _ := primitive.ObjectIDFromHex(userID.(string))
    
    ctx := context.Background()
    result, err := h.collection.UpdateOne(
        ctx,
        bson.M{
            "_id": id,
            "user_id": objID,
        },
        bson.M{
            "$set": bson.M{
                "metadata": metadata,
                "updated_at": time.Now(),
            },
        },
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update media"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
        return
    }

    c.Status(http.StatusOK)
}

// generateURL creates a public URL for the file
func (h *MediaHandler) generateURL(filePath string) string {
    relativePath := strings.TrimPrefix(filePath, "uploads/")
    return h.baseURL + "/media/" + relativePath
}
