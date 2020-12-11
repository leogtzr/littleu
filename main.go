package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/go-redis/redis"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"

	redisSession "github.com/gin-contrib/sessions/redis"

	"encoding/gob"
)

func init() {
	var err error
	envConfig, err = readConfig("config.env", ".", map[string]interface{}{
		"dbengine": "memory",
		"port":     "8080",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	serverPort = envConfig.GetString("port")

	//Initializing redis
	dsn := envConfig.GetString("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	ctx = context.TODO()

	// Initialize DB:
	urlDAO = factoryURLDao(envConfig.GetString("dbengine"), envConfig)
	userDAO = factoryUserDAO(envConfig.GetString("dbengine"), envConfig)

	gob.Register(&User{})
	gob.Register(&UserPostgresql{})
}

func main() {

	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	router = gin.Default()
	// store := cookie.NewStore([]byte(envConfig.GetString("SESSION_SECRET")))
	//Initializing redis
	dsn := envConfig.GetString("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}

	store, _ := redisSession.NewStore(10, "tcp", dsn, "", []byte(envConfig.GetString("SESSION_SECRET")))
	router.Use(sessions.Sessions("sid", store))

	router.Static("/assets", "./assets")

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	// Start serving the applications
	router.Run(net.JoinHostPort("", serverPort))
}
