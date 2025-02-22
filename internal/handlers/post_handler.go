package handlers

import (
    "context"
    "net/http"
    "time"
    "path/filepath"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"

    "go-blog-platform/internal/models"
    "go-blog-platform/internal/services"
    "mime/multipart"
)

type PostHandler struct {
    collection   *mongo.Collection
    mediaService *services.MediaService
}

type CreatePostRequest struct {
    Title        string   `json:"title" binding:"required"`
    Content      string   `json:"content" binding:"required"`
    Tags         []string `json:"tags"`
    Status       string   `json:"status" binding:"required,oneof=published draft"`
    FeaturedFile *multipart.FileHeader `form:"featured_image"`
    GalleryFiles []*multipart.FileHeader `form:"gallery[]"`
}

func NewPostHandler(db *mongo.Database, mediaService *services.MediaService) *PostHandler {
    return &PostHandler{
        collection:   db.Collection("posts"),
        mediaService: mediaService,
    }
}

func (h *PostHandler) List(c *gin.Context) {
    ctx := context.Background()
    cursor, err := h.collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(ctx)

    var posts []models.Post
    if err := cursor.All(ctx, &posts); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) Create(c *gin.Context) {
    var req CreatePostRequest
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get user ID from context
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    objID, _ := primitive.ObjectIDFromHex(userID.(string))
    post := models.Post{
        ID:        primitive.NewObjectID(),
        Title:     req.Title,
        Content:   req.Content,
        AuthorID:  objID,
        Status:    req.Status,
        Tags:      req.Tags,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Handle featured image upload
    if req.FeaturedFile != nil {
        if err := h.mediaService.ValidateFile(req.FeaturedFile); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid featured image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.FeaturedFile, userID.(string))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save featured image"})
            return
        }

        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    objID,
            FileName:  req.FeaturedFile.Filename,
            FileType:  filepath.Ext(req.FeaturedFile.Filename),
            MimeType:  req.FeaturedFile.Header.Get("Content-Type"),
            Size:      req.FeaturedFile.Size,
            Path:      filePath,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        for _, thumb := range thumbnails {
            media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                Size: thumb.Size,
                Path: thumb.Path,
            })
        }

        post.FeaturedImage = media
    }

    // Handle gallery images upload
    if len(req.GalleryFiles) > 0 {
        for _, file := range req.GalleryFiles {
            if err := h.mediaService.ValidateFile(file); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery image: " + err.Error()})
                return
            }

            filePath, thumbnails, err := h.mediaService.SaveFile(file, userID.(string))
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save gallery image"})
                return
            }

            media := models.Media{
                ID:        primitive.NewObjectID(),
                UserID:    objID,
                FileName:  file.Filename,
                FileType:  filepath.Ext(file.Filename),
                MimeType:  file.Header.Get("Content-Type"),
                Size:      file.Size,
                Path:      filePath,
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            }

            for _, thumb := range thumbnails {
                media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                    Size: thumb.Size,
                    Path: thumb.Path,
                })
            }

            post.Gallery = append(post.Gallery, media)
        }
    }

    ctx := context.Background()
    _, err = h.collection.InsertOne(ctx, post)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
        return
    }

    c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) Get(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
        return
    }

    ctx := context.Background()
    var post models.Post
    err = h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Update(c *gin.Context) {
    id := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
        return
    }

    var req CreatePostRequest
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Get existing post
    ctx := context.Background()
    var existingPost models.Post
    err = h.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existingPost)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    // Update basic fields
    update := bson.M{
        "$set": bson.M{
            "title":      req.Title,
            "content":    req.Content,
            "status":     req.Status,
            "tags":       req.Tags,
            "updated_at": time.Now(),
        },
    }

    // Handle featured image update
    if req.FeaturedFile != nil {
        // Delete old featured image if exists
        if existingPost.FeaturedImage != nil {
            _ = h.mediaService.DeleteFile(existingPost.FeaturedImage.Path)
        }

        // Upload new featured image
        if err := h.mediaService.ValidateFile(req.FeaturedFile); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid featured image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.FeaturedFile, userID.(string))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save featured image"})
            return
        }

        objID, _ := primitive.ObjectIDFromHex(userID.(string))
        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    objID,
            FileName:  req.FeaturedFile.Filename,
            FileType:  filepath.Ext(req.FeaturedFile.Filename),
            MimeType:  req.FeaturedFile.Header.Get("Content-Type"),
            Size:      req.FeaturedFile.Size,
            Path:      filePath,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        for _, thumb := range thumbnails {
            media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                Size: thumb.Size,
                Path: thumb.Path,
            })
        }

        update["$set"].(bson.M)["featured_image"] = media
    }

    // Handle gallery updates
    if len(req.GalleryFiles) > 0 {
        // Delete old gallery images
        for _, media := range existingPost.Gallery {
            _ = h.mediaService.DeleteFile(media.Path)
        }

        // Upload new gallery images
        var gallery []models.Media
        for _, file := range req.GalleryFiles {
            if err := h.mediaService.ValidateFile(file); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery image: " + err.Error()})
                return
            }

            filePath, thumbnails, err := h.mediaService.SaveFile(file, userID.(string))
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save gallery image"})
                return
            }

            objID, _ := primitive.ObjectIDFromHex(userID.(string))
            media := models.Media{
                ID:        primitive.NewObjectID(),
                UserID:    objID,
                FileName:  file.Filename,
                FileType:  filepath.Ext(file.Filename),
                MimeType:  file.Header.Get("Content-Type"),
                Size:      file.Size,
                Path:      filePath,
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            }

            for _, thumb := range thumbnails {
                media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                    Size: thumb.Size,
                    Path: thumb.Path,
                })
            }

            gallery = append(gallery, media)
        }

        update["$set"].(bson.M)["gallery"] = gallery
    }

    result, err := h.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    // Get updated post
    var updatedPost models.Post
    err = h.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedPost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated post"})
        return
    }

    c.JSON(http.StatusOK, updatedPost)
}

func (h *PostHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
        return
    }

    ctx := context.Background()
    
    // Get post to delete associated media
    var post models.Post
    err = h.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&post)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
        return
    }

    // Delete associated media
    if post.FeaturedImage != nil {
        _ = h.mediaService.DeleteFile(post.FeaturedImage.Path)
    }
    for _, media := range post.Gallery {
        _ = h.mediaService.DeleteFile(media.Path)
    }

    // Delete post
    result, err := h.collection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    c.Status(http.StatusNoContent)
}

// ListDrafts returns all draft posts for the current user
func (h *PostHandler) ListDrafts(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    ctx := context.Background()
    cursor, err := h.collection.Find(ctx, bson.M{
        "user_id": userID.(string),
        "published": false,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drafts"})
        return
    }
    defer cursor.Close(ctx)

    var drafts []models.Post
    if err := cursor.All(ctx, &drafts); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode drafts"})
        return
    }

    c.JSON(http.StatusOK, drafts)
}

// CreateDraft creates a new draft post
func (h *PostHandler) CreateDraft(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var post models.Post
    if err := c.ShouldBindJSON(&post); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post.ID = primitive.NewObjectID()
    post.UserID = userID.(string)
    post.Published = false // Ensure it's marked as unpublished
    post.CreatedAt = time.Now()
    post.UpdatedAt = time.Now()

    ctx := context.Background()
    _, err := h.collection.InsertOne(ctx, post)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create draft"})
        return
    }

    c.JSON(http.StatusCreated, post)
}
