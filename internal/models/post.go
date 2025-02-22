package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title     string            `bson:"title" json:"title"`
	Content   string            `bson:"content" json:"content"`
	Slug      string            `bson:"slug" json:"slug"`
	Published bool              `bson:"published" json:"published"`
	UserID    string            `bson:"user_id" json:"user_id"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}
