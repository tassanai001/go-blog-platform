package models

import (
	"context"
	"errors"
	"time"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go-blog-platform/internal/constants"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"` // "-" means this field won't be included in JSON responses
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// ValidateRole checks if the role is valid
func (u *User) ValidateRole() error {
	switch u.Role {
	case constants.RoleAdmin, constants.RoleAuthor, constants.RoleReader:
		return nil
	default:
		return errors.New("invalid role")
	}
}

// HasPermission checks if the user has the required role permission
func (u *User) HasPermission(requiredRole string) bool {
	allowedRoles, exists := constants.RoleHierarchy[u.Role]
	if !exists {
		return false
	}
	
	for _, role := range allowedRoles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// BeforeInsert is called before inserting a new user
func (u *User) BeforeInsert(ctx context.Context) error {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
	
	// Generate new ObjectID if not set
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}

	// Set default role if not specified
	if u.Role == "" {
		u.Role = constants.RoleReader
	}

	// Validate role
	if err := u.ValidateRole(); err != nil {
		return err
	}

	// Hash password if not already hashed
	if len(u.Password) > 0 && len(u.Password) < 60 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	
	return nil
}

func (u *User) BeforeUpdate(ctx context.Context) error {
	u.UpdatedAt = time.Now()
	return nil
}

// ComparePassword checks if the provided password matches the hashed password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}