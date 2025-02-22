package handlers

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"

    "go-blog-platform/internal/models"
    "go-blog-platform/internal/constants"
    "go-blog-platform/internal/services"
)

type UserHandler struct {
    collection          *mongo.Collection
    resetTokens         *mongo.Collection
    jwtSecret          []byte
    emailService       *services.EmailService
    baseURL           string
}

func NewUserHandler(db *mongo.Database, jwtSecret string, emailService *services.EmailService, baseURL string) *UserHandler {
    return &UserHandler{
        collection:    db.Collection("users"),
        resetTokens:   db.Collection("password_reset_tokens"),
        jwtSecret:     []byte(jwtSecret),
        emailService:  emailService,
        baseURL:      baseURL,
    }
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Role     string `json:"role"`
}

type UpdateRoleRequest struct {
    Role string `json:"role" binding:"required"`
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
    case constants.RoleAdmin, constants.RoleAuthor, constants.RoleReader:
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
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if user already exists
    ctx := context.Background()
    var existingUser models.User
    err := h.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        return
    }
    if err != mongo.ErrNoDocuments {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Create new user
    user := models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
        Role:     req.Role, // Will be set to default "reader" in BeforeInsert if empty
    }

    if err := user.BeforeInsert(ctx); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    _, err = h.collection.InsertOne(ctx, user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "User registered successfully",
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
            "role":     user.Role,
        },
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
