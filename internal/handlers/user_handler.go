package handlers

import (
    "context"
    "fmt"
    "log"
    "mime/multipart"
    "net/http"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"

    "go-blog-platform/internal/models"
    "go-blog-platform/internal/services"
)

type UserHandler struct {
    collection          *mongo.Collection
    resetTokens         *mongo.Collection
    jwtSecret          []byte
    emailService       *services.EmailService
    mediaService       *services.MediaService
    baseURL           string
}

func NewUserHandler(db *mongo.Database, jwtSecret string, emailService *services.EmailService, mediaService *services.MediaService, baseURL string) *UserHandler {
    return &UserHandler{
        collection:    db.Collection("users"),
        resetTokens:   db.Collection("password_reset_tokens"),
        jwtSecret:     []byte(jwtSecret),
        emailService:  emailService,
        mediaService:  mediaService,
        baseURL:      baseURL,
    }
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
    Username    string               `form:"username" binding:"required"`
    Password    string               `form:"password" binding:"required"`
    Email       string               `form:"email" binding:"required,email"`
    FullName    string               `form:"full_name" binding:"required"`
    Avatar      *multipart.FileHeader `form:"avatar"`
    CoverImage  *multipart.FileHeader `form:"cover_image"`
    Bio         string               `form:"bio"`
    Location    string               `form:"location"`
    Website     string               `form:"website"`
    SocialLinks models.SocialLinks   `form:"social_links"`
    Role        string               `json:"role"`
}

type UpdateRoleRequest struct {
    Role string `json:"role" binding:"required"`
}

type UpdateProfileRequest struct {
    FullName    string               `form:"full_name"`
    Avatar      *multipart.FileHeader `form:"avatar"`
    CoverImage  *multipart.FileHeader `form:"cover_image"`
    Bio         string               `form:"bio"`
    Location    string               `form:"location"`
    Website     string               `form:"website"`
    SocialLinks models.SocialLinks   `form:"social_links"`
}

type RequestPasswordResetRequest struct {
    Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
    Token    string `json:"token" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) ListUsers(c *gin.Context) {
    ctx := context.Background()
    cursor, err := h.collection.Find(ctx, bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    defer cursor.Close(ctx)

    var users []models.User
    if err := cursor.All(ctx, &users); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
        return
    }

    c.JSON(http.StatusOK, users)
}

// UpdateUserRole updates a user's role (admin only)
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
    userID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var req UpdateRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate the new role
    switch req.Role {
    case "admin", "author", "reader":
        // Valid role
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
        return
    }

    ctx := context.Background()
    result, err := h.collection.UpdateOne(
        ctx,
        bson.M{"_id": userID},
        bson.M{"$set": bson.M{"role": req.Role, "updated_at": time.Now()}},
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// DeleteUser deletes a user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID, err := primitive.ObjectIDFromHex(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    ctx := context.Background()
    result, err := h.collection.DeleteOne(ctx, bson.M{"_id": userID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if username already exists
    var existingUser models.User
    err := h.collection.FindOne(context.Background(), bson.M{"username": req.Username}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
        return
    }

    // Check if email already exists
    err = h.collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    userID := primitive.NewObjectID()
    now := time.Now()

    // Create profile with media handling
    profile := models.Profile{
        ID:          primitive.NewObjectID(),
        UserID:      userID,
        FullName:    req.FullName,
        Bio:         req.Bio,
        Location:    req.Location,
        Website:     req.Website,
        SocialLinks: req.SocialLinks,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // Handle avatar upload
    if req.Avatar != nil {
        if err := h.mediaService.ValidateFile(req.Avatar); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid avatar image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.Avatar, userID.Hex())
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
            return
        }

        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    userID,
            FileName:  req.Avatar.Filename,
            FileType:  filepath.Ext(req.Avatar.Filename),
            MimeType:  req.Avatar.Header.Get("Content-Type"),
            Size:      req.Avatar.Size,
            Path:      filePath,
            CreatedAt: now,
            UpdatedAt: now,
        }

        for _, thumb := range thumbnails {
            media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                Size: thumb.Size,
                Path: thumb.Path,
            })
        }

        profile.Avatar = media
    }

    // Handle cover image upload
    if req.CoverImage != nil {
        if err := h.mediaService.ValidateFile(req.CoverImage); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cover image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.CoverImage, userID.Hex())
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover image"})
            return
        }

        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    userID,
            FileName:  req.CoverImage.Filename,
            FileType:  filepath.Ext(req.CoverImage.Filename),
            MimeType:  req.CoverImage.Header.Get("Content-Type"),
            Size:      req.CoverImage.Size,
            Path:      filePath,
            CreatedAt: now,
            UpdatedAt: now,
        }

        for _, thumb := range thumbnails {
            media.Thumbnails = append(media.Thumbnails, models.Thumbnail{
                Size: thumb.Size,
                Path: thumb.Path,
            })
        }

        profile.CoverImage = media
    }

    // Create user
    user := models.User{
        ID:        userID,
        Username:  req.Username,
        Email:     req.Email,
        Password:  string(hashedPassword),
        Profile:   profile,
        Role:      req.Role,
        CreatedAt: now,
        UpdatedAt: now,
    }

    _, err = h.collection.InsertOne(context.Background(), user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "id":       user.ID,
        "username": user.Username,
        "profile":  user.Profile,
    })
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    var req UpdateProfileRequest
    if err := c.ShouldBind(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    objID, _ := primitive.ObjectIDFromHex(userID.(string))

    // Get existing user
    var user models.User
    err := h.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Update basic profile fields
    update := bson.M{
        "$set": bson.M{
            "profile.full_name":     req.FullName,
            "profile.bio":           req.Bio,
            "profile.location":      req.Location,
            "profile.website":       req.Website,
            "profile.social_links":  req.SocialLinks,
            "profile.updated_at":    time.Now(),
        },
    }

    // Handle avatar update
    if req.Avatar != nil {
        // Delete old avatar if exists
        if user.Profile.Avatar != nil {
            _ = h.mediaService.DeleteFile(user.Profile.Avatar.Path)
        }

        if err := h.mediaService.ValidateFile(req.Avatar); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid avatar image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.Avatar, userID.(string))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
            return
        }

        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    objID,
            FileName:  req.Avatar.Filename,
            FileType:  filepath.Ext(req.Avatar.Filename),
            MimeType:  req.Avatar.Header.Get("Content-Type"),
            Size:      req.Avatar.Size,
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

        update["$set"].(bson.M)["profile.avatar"] = media
    }

    // Handle cover image update
    if req.CoverImage != nil {
        // Delete old cover image if exists
        if user.Profile.CoverImage != nil {
            _ = h.mediaService.DeleteFile(user.Profile.CoverImage.Path)
        }

        if err := h.mediaService.ValidateFile(req.CoverImage); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cover image: " + err.Error()})
            return
        }

        filePath, thumbnails, err := h.mediaService.SaveFile(req.CoverImage, userID.(string))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cover image"})
            return
        }

        media := &models.Media{
            ID:        primitive.NewObjectID(),
            UserID:    objID,
            FileName:  req.CoverImage.Filename,
            FileType:  filepath.Ext(req.CoverImage.Filename),
            MimeType:  req.CoverImage.Header.Get("Content-Type"),
            Size:      req.CoverImage.Size,
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

        update["$set"].(bson.M)["profile.cover_image"] = media
    }

    result, err := h.collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
        return
    }

    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Get updated user
    var updatedUser models.User
    err = h.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&updatedUser)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":       updatedUser.ID,
        "username": updatedUser.Username,
        "profile":  updatedUser.Profile,
    })
}

func (h *UserHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find user by email
    ctx := context.Background()
    var user models.User
    err := h.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Check password
    if err := user.ComparePassword(req.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID.Hex(),
        "email":   user.Email,
        "role":    user.Role,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(h.jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "token": tokenString,
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
            "role":     user.Role,
        },
    })
}

// RequestPasswordReset initiates the password reset process
func (h *UserHandler) RequestPasswordReset(c *gin.Context) {
    var req RequestPasswordResetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find user by email
    ctx := context.Background()
    var user models.User
    err := h.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        // Don't reveal whether the email exists
        c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent"})
        return
    }
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Generate reset token
    token := primitive.NewObjectID().Hex()
    resetToken := models.PasswordResetToken{
        ID:        primitive.NewObjectID(),
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(1 * time.Hour),
        Used:      false,
        CreatedAt: time.Now(),
    }

    // Save reset token
    _, err = h.resetTokens.InsertOne(ctx, resetToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
        return
    }

    // Generate reset link
    resetLink := fmt.Sprintf("%s/reset-password?token=%s", h.baseURL, token)

    // Send reset email
    err = h.emailService.SendPasswordResetEmail(user.Email, resetLink)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent"})
}

// ResetPassword handles the actual password reset
func (h *UserHandler) ResetPassword(c *gin.Context) {
    var req ResetPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := context.Background()

    // Find and validate reset token
    var resetToken models.PasswordResetToken
    err := h.resetTokens.FindOne(ctx, bson.M{
        "token": req.Token,
        "used":  false,
        "expires_at": bson.M{"$gt": time.Now()},
    }).Decode(&resetToken)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Update user's password
    _, err = h.collection.UpdateOne(
        ctx,
        bson.M{"_id": resetToken.UserID},
        bson.M{"$set": bson.M{
            "password":   string(hashedPassword),
            "updated_at": time.Now(),
        }},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }

    // Mark reset token as used
    _, err = h.resetTokens.UpdateOne(
        ctx,
        bson.M{"_id": resetToken.ID},
        bson.M{"$set": bson.M{"used": true}},
    )
    if err != nil {
        // Log this error but don't return it to the user since the password was updated
        log.Printf("Failed to mark reset token as used: %v", err)
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}
