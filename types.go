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

// URLStat ...
type URLStat struct {
	ShortID int     `json:"id"`
	Url     string	`json:"url"`
}

// URLStatFull is basically a URLStat but instead of the short ID, it has the short URL corresponding
// to the short ID value.
type URLStatFull struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

// StatsMongo ...
type StatsMongo struct {
	ID			primitive.ObjectID	`bson:"_id"`
	CreatedAt	time.Time          	`bson:"created_at"`
	ShortID		int					`bson:"shortid"`
	UserID		int					`bson:"user_id"`
	Headers		map[string][]string `bson:"req_info"`
}

// StatsPostgresql ...
type StatsPostgresql struct {
	ID        	int
	CreatedAt 	time.Time
	ShortID   	int
	UserID		int
}

// StatsInMemory ...
type StatsInMemory struct {
	CreatedAt	time.Time
	ShortID		int
	Headers		map[string][]string
}

// StatsHeadersPostgresql ...
type StatsHeadersPostgresql struct {
	/*
		id serial PRIMARY KEY NOT NULL,
		name varchar(150) NOT NULL,
		value varchar(500) NOT NULL,
		stat_id int NOT NULL,
	    constraint fk_stats_headers
	        foreign key (stat_id)
	        REFERENCES stats (id)
	 */

}
