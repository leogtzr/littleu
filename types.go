package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// URL ...
type URL struct {
	URL string `form:"url"`
}

// URLChange ...
type URLChange struct {
	ShortURL string `form:"url"`
	NewURL   string `form:"new_url"`
}

// User ...
type User struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	User      string             `bson:"user"`
	Password  string             `bson:"password"`
}
