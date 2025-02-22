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

type PostHandler struct {
    collection *mongo.Collection
}

func NewPostHandler(db *mongo.Database) *PostHandler {
    return &PostHandler{
        collection: db.Collection("posts"),
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
    var post models.Post
    if err := c.ShouldBindJSON(&post); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post.ID = primitive.NewObjectID()
    post.CreatedAt = time.Now()
    post.UpdatedAt = time.Now()

    ctx := context.Background()
    _, err := h.collection.InsertOne(ctx, post)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
        return
    }

    var post models.Post
    if err := c.ShouldBindJSON(&post); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post.UpdatedAt = time.Now()

    ctx := context.Background()
    result, err := h.collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{
            "$set": bson.M{
                "title":      post.Title,
                "content":    post.Content,
                "slug":       post.Slug,
                "published":  post.Published,
                "updated_at": post.UpdatedAt,
            },
        },
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    post.ID = id
    c.JSON(http.StatusOK, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
        return
    }

    ctx := context.Background()
    result, err := h.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    c.Status(http.StatusNoContent)
}
