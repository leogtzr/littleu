package main

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	collection *mongo.Collection
	ctx        = context.TODO()

	mu sync.RWMutex

	// TODO: remove the following:
	// dummy auth user:
	dummyUser = User{
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
