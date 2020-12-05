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

// @TODO: this might be changed
// User ...
type User struct {
	ID uint64 `json:"id,omitempty"`
	// ID       uint64
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewUser ...
type NewUser struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	User      string             `bson:"user"`
	Password  string             `bson:"password"`
}
