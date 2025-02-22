package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
    FullName    string            `bson:"full_name" json:"full_name"`
    Bio         string            `bson:"bio" json:"bio"`
    Avatar      string            `bson:"avatar" json:"avatar"`
    Website     string            `bson:"website" json:"website"`
    Location    string            `bson:"location" json:"location"`
    SocialLinks SocialLinks       `bson:"social_links" json:"social_links"`
    CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type SocialLinks struct {
    Twitter   string `bson:"twitter" json:"twitter"`
    Facebook  string `bson:"facebook" json:"facebook"`
    LinkedIn  string `bson:"linkedin" json:"linkedin"`
    GitHub    string `bson:"github" json:"github"`
    Instagram string `bson:"instagram" json:"instagram"`
}
