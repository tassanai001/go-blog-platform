package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	FullName    string            `bson:"full_name" json:"full_name"`
	Bio         string            `bson:"bio,omitempty" json:"bio,omitempty"`
	Avatar      *Media            `bson:"avatar,omitempty" json:"avatar,omitempty"`
	CoverImage  *Media            `bson:"cover_image,omitempty" json:"cover_image,omitempty"`
	Location    string            `bson:"location,omitempty" json:"location,omitempty"`
	Website     string            `bson:"website,omitempty" json:"website,omitempty"`
	SocialLinks SocialLinks       `bson:"social_links,omitempty" json:"social_links,omitempty"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type SocialLinks struct {
	Twitter   string `bson:"twitter,omitempty" json:"twitter,omitempty"`
	Facebook  string `bson:"facebook,omitempty" json:"facebook,omitempty"`
	LinkedIn  string `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
	GitHub    string `bson:"github,omitempty" json:"github,omitempty"`
	Instagram string `bson:"instagram,omitempty" json:"instagram,omitempty"`
}
