package main

import (
	"sync"

	"github.com/go-redis/redis"
)

var (
	mu sync.RWMutex

	// TODO: remove the following:
	// dummy auth user:
	user = User{
		ID:       1,
		Username: "username",
		Password: "password",
	}

	redisClient *redis.Client
)
