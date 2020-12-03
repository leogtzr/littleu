package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
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

var (
	router     *gin.Engine
	envConfig  *viper.Viper
	dao        *DBHandler
	serverPort string
)
