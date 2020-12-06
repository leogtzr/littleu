package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var (
	mu sync.RWMutex

	redisClient *redis.Client
)

var (
	router    *gin.Engine
	envConfig *viper.Viper
	urlDAO    *URLDao
	userDAO   *UserDAO

	serverPort string
)
