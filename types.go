package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// URL ...
type URL struct {
	URL string `form:"url"`
}

// URLDocument ...
type URLDocument struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	ShortID   int                `bson:"shortid"`
	URL       string             `bson:"url"`
}

// URLChange ...
type URLChange struct {
	ShortURL string `form:"url"`
	NewURL   string `form:"new_url"`
}

// User ...
type User struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	User      string             `json:"user" bson:"user"`
	Password  string             `json:"password" bson:"password"`
	// IDP       int
}

// UserPostgresql ...
type UserPostgresql struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	User      string
	Password  string
}
