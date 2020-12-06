// TODO: fix middleware to know if the user is already logged or not.
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
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

	// Initialize DB:
	dao = factory(envConfig.GetString("dbengine"))

	//Initializing redis
	dsn := envConfig.GetString("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	mongoClientOptions := options.Client().ApplyURI(envConfig.GetString("MONGO_URI"))
	client, err := mongo.Connect(ctx, mongoClientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("littleu").Collection("user")
}

func main() {

	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	router = gin.Default()

	router.Static("/assets", "./assets")

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	// Start serving the applications
	router.Run(net.JoinHostPort("", serverPort))
}
