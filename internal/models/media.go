package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Media struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
    FileName    string            `bson:"file_name" json:"file_name"`
    FileType    string            `bson:"file_type" json:"file_type"`
    MimeType    string            `bson:"mime_type" json:"mime_type"`
    Size        int64             `bson:"size" json:"size"`
    Path        string            `bson:"path" json:"path"`
    URL         string            `bson:"url" json:"url"`
    Thumbnails  []Thumbnail       `bson:"thumbnails" json:"thumbnails,omitempty"`
    Metadata    MediaMetadata     `bson:"metadata" json:"metadata,omitempty"`
    CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type Thumbnail struct {
    Size   string `bson:"size" json:"size"`     // e.g., "small", "medium", "large"
    Width  int    `bson:"width" json:"width"`
    Height int    `bson:"height" json:"height"`
    Path   string `bson:"path" json:"path"`
    URL    string `bson:"url" json:"url"`
}

type MediaMetadata struct {
    Width       int      `bson:"width,omitempty" json:"width,omitempty"`
    Height      int      `bson:"height,omitempty" json:"height,omitempty"`
    Title       string   `bson:"title,omitempty" json:"title,omitempty"`
    Description string   `bson:"description,omitempty" json:"description,omitempty"`
    AltText     string   `bson:"alt_text,omitempty" json:"alt_text,omitempty"`
    Tags        []string `bson:"tags,omitempty" json:"tags,omitempty"`
}
