package models

import (
	"context"
	"errors"
	"go-blog-platform/internal/constants"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
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

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	if u.Password == "" {
		return errors.New("password is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword compares the provided password with the user's hashed password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
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
		if err := u.HashPassword(); err != nil {
			return err
		}
	}

	return nil
}

func (u *User) BeforeUpdate(ctx context.Context) error {
	u.UpdatedAt = time.Now()
	return nil
}
