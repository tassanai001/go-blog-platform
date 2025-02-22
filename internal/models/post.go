package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title        string            `bson:"title" json:"title"`
	Content      string            `bson:"content" json:"content"`
	AuthorID     primitive.ObjectID `bson:"author_id" json:"author_id"`
	FeaturedImage *Media            `bson:"featured_image,omitempty" json:"featured_image,omitempty"`
	Gallery      []Media           `bson:"gallery,omitempty" json:"gallery,omitempty"`
	Status       string            `bson:"status" json:"status"` // published, draft
	Tags         []string          `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt    time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at" json:"updated_at"`
}
