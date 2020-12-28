package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

type randGenSrc struct{}

func (s *randGenSrc) Seed(int64) {}

func (s *randGenSrc) Uint64() (value uint64) {
	_ = binary.Read(rand.Reader, binary.BigEndian, &value)

	return value
}

// URLDao ...
type URLDao interface {
	save(url URL, user *interface{}) (int, error)
	update(id int, oldURL, newURL URL) (int, error)
	findByID(id int) (URL, error)
	findAllByUser(id *interface{}) ([]URLStat, error)
}

// UserDAO ....
type UserDAO interface {
	addUser(username, password string) (interface{}, error)
	userExists(username string) (bool, error)
	findByUsername(username string) (interface{}, error)
	validateUserAndPassword(username, password string) (bool, error)
	findAll() ([]interface{}, error)
}

// StatsDAO ...
type StatsDAO interface {
	save(shortURL string, headers *map[string][]string, user *interface{}) (int, error)
	findByShortID(id int) ([]interface{}, error)
	findAllByUser(user *interface{}) ([]interface{}, error)
}

func factoryStatsDao(mongoClient *mongo.Client, config *viper.Viper) *StatsDAO {
	var dao StatsDAO

	engine := config.GetString("dbengine")

	switch engine {
	case "memory":
		dao = StatsDAOMemoryImpl{
			db: map[int][]StatsInMemory{},
		}
	case "mongo":
		var collection *mongo.Collection
		collection = mongoClient.Database("littleu").Collection("stats")
		dao = StatsMongoImpl{
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

		dao = StatsPostgresqlImpl{
			db,
		}
	default:
		log.Fatalf("error: wrong engine: %s", engine)

		return nil
	}

	return &dao
}

func factoryURLDao(mongoClient *mongo.Client, config *viper.Viper) *URLDao {
	var dao URLDao

	engine := config.GetString("dbengine")

	switch engine {
	case "memory":
		dao = InMemoryURLDAOImpl{
			DB: &memoryDB{
				db: map[int]string{},
			},
		}
	case "mongo":
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

func factoryUserDAO(mongoClient *mongo.Client, config *viper.Viper) *UserDAO {
	var userDAO UserDAO

	engine := config.GetString("dbengine")

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
