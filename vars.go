package main

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	mongoClientOptions *options.ClientOptions
	mongoClient        *mongo.Client
	ctx                context.Context
)
