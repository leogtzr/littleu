package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type randGenSrc struct{}

func (s *randGenSrc) Seed(seed int64) {}

func (s *randGenSrc) Uint64() (value uint64) {
	_ = binary.Read(rand.Reader, binary.BigEndian, &value)

	return value
}

// URLDao ...
type URLDao interface {
	save(url URL, user *interface{}) (int, error)
	update(id int, oldURL, newURL URL) (int, error)
	findByID(id int) (URL, error)
	findAll() (map[int]string, error)
}

// UserDAO ....
type UserDAO interface {
	addUser(username, password string) (interface{}, error)
	userExists(username string) (bool, error)
	findByUsername(username string) (interface{}, error)
	validateUserAndPassword(username, password string) (bool, error)
	findAll() ([]interface{}, error)
}

func factoryURLDao(engine string, config *viper.Viper) *URLDao {
	var dao URLDao

	switch engine {
	case "memory":
		dao = InMemoryURLDAOImpl{
			DB: &memoryDB{
				db: map[int]string{},
			},
		}
	case "mongo":
		var err error

		mongoClientOptions = options.Client().ApplyURI(config.GetString("MONGO_URI"))

		mongoClient, err = mongo.Connect(ctx, mongoClientOptions)
		if err != nil {
			log.Fatal(err)
		}

		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}

		var collection *mongo.Collection
		collection = mongoClient.Database("littleu").Collection("url")
		dao = MongoDBURLDAOImpl{
			collection: collection,
			ctx:        ctx,
		}

	case "postgresql":
		dsn := config.GetString("POSTGRES_DSN")
		if dsn == "" {
			log.Fatalf("POSTGRES_DSN environtment variable is not set")
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatal(err)
		}

		dao = PostgresqlURLDAOImpl{
			db,
		}
	default:
		log.Fatalf("error: wrong engine: %s", engine)

		return nil
	}

	return &dao
}

func factoryUserDAO(engine string, config *viper.Viper) *UserDAO {
	var userDAO UserDAO

	switch engine {
	case "memory":
		userDAO = InMemoryUserDAOImpl{
			db:       map[string]UserInMemory{},
			rndIDGen: randGenSrc{},
		}
	case "mongo":
		var collection *mongo.Collection
		collection = mongoClient.Database("littleu").Collection("user")

		userDAO = MongoUserDaoImpl{
			collection: collection,
			ctx:        ctx,
		}
	case "postgresql":
		dsn := config.GetString("POSTGRES_DSN")
		if dsn == "" {
			log.Fatalf("POSTGRES_DSN environtment variable is not set")
		}

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatal(err)
		}

		userDAO = PostgresqlUserImpl{
			db,
		}

	default:
		log.Fatalf("error: wrong engine: %s", engine)

		return nil
	}

	return &userDAO
}
