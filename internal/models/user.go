package models

import (
	"context"
	"errors"
	"fmt"
	"time"
	"golang.org/x/crypto/bcrypt"
	"go-blog-platform/internal/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Role      string             `bson:"role" json:"role"`
	Profile   Profile            `bson:"profile" json:"profile"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	FullName    string             `bson:"full_name" json:"full_name"`
	Bio         string             `bson:"bio,omitempty" json:"bio,omitempty"`
	Avatar      *Media             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CoverImage  *Media             `bson:"cover_image,omitempty" json:"cover_image,omitempty"`
	Location    string             `bson:"location,omitempty" json:"location,omitempty"`
	Website     string             `bson:"website,omitempty" json:"website,omitempty"`
	SocialLinks SocialLinks        `bson:"social_links,omitempty" json:"social_links,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type Media struct {
	// Add media fields here
}

type SocialLinks struct {
	Twitter   string `bson:"twitter,omitempty" json:"twitter,omitempty"`
	Facebook  string `bson:"facebook,omitempty" json:"facebook,omitempty"`
	LinkedIn  string `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
	GitHub    string `bson:"github,omitempty" json:"github,omitempty"`
	Instagram string `bson:"instagram,omitempty" json:"instagram,omitempty"`
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