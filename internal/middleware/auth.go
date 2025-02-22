package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "go-blog-platform/internal/constants"
)

func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        c.Set("user_id", claims["user_id"])
        c.Set("email", claims["email"])
        c.Set("role", claims["role"])

        c.Next()
    }
}

// RequireRole middleware checks if the user has the required role or higher
func RequireRole(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        allowedRoles, exists := constants.RoleHierarchy[userRole.(string)]
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
            c.Abort()
            return
        }

        hasPermission := false
        for _, role := range allowedRoles {
            if role == requiredRole {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// IsAdmin is a convenience middleware for admin-only routes
func IsAdmin() gin.HandlerFunc {
    return RequireRole(constants.RoleAdmin)
}

// IsAuthorOrAdmin is a convenience middleware for routes that require author or admin privileges
func IsAuthorOrAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        allowedRoles, exists := constants.RoleHierarchy[userRole.(string)]
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
            c.Abort()
            return
        }

        hasPermission := false
        for _, role := range allowedRoles {
            if role == constants.RoleAdmin || role == constants.RoleAuthor {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
