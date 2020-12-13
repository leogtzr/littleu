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
	UserID    primitive.ObjectID `bson:"user_id"`
}

// URLChange ...
type URLChange struct {
	ShortURL string `form:"url"`
	NewURL   string `form:"new_url"`
}

// UserMongo ...
type UserMongo struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	User      string             `json:"user" bson:"user"`
	Password  string             `json:"password" bson:"password"`
}

// UserPostgresql ...
type UserPostgresql struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	User      string
	Password  string
}

// UserInMemory ...
type UserInMemory struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	User      string
	Password  string
}
