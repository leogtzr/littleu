package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// URLDao ...
type URLDao interface {
	save(u URL) (int, error)
	update(ID int, oldURL, newURL URL) (int, error)
	findAll() (map[int]string, error)
	findByID(ID int) (URL, error)
}

// TODO: finish this.

// UserDAO ....
type UserDAO interface {
	save(user *User) (primitive.ObjectID, error)
	userExists(username string) (bool, error)
	findByUsername(username string) (User, error)
	findAll() ([]User, error)
}

type memoryDB struct {
	db            map[int]string
	autoIncrement int
}

// InMemoryImpl ...
type InMemoryImpl struct {
	DB *memoryDB
}

// MongoUserDaoImpl ...
type MongoUserDaoImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func factoryURLDao(engine string) *URLDao {
	var dao URLDao
	switch engine {
	case "memory":
		dao = InMemoryImpl{
			DB: &memoryDB{
				db: map[int]string{},
			},
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
	case "mongo":
		mongoClientOptions := options.Client().ApplyURI(config.GetString("MONGO_URI"))
		ctx := context.TODO()
		client, err := mongo.Connect(ctx, mongoClientOptions)
		if err != nil {
			log.Fatal(err)
		}

		err = client.Ping(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}

		var collection *mongo.Collection

		collection = client.Database("littleu").Collection("user")

		userDAO = MongoUserDaoImpl{
			collection,
			ctx,
		}
	}

	return &userDAO
}

func (im InMemoryImpl) save(u URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	im.DB.autoIncrement++
	id := im.DB.autoIncrement
	im.DB.db[id] = u.URL

	return id, nil
}

func (im InMemoryImpl) findAll() (map[int]string, error) {
	return im.DB.db, nil
}

func (im InMemoryImpl) findByID(ID int) (URL, error) {
	u, found := im.DB.db[ID]
	if found {
		url := URL{
			URL: u,
		}
		return url, nil
	}
	return URL{}, fmt.Errorf("no url found for: %d", ID)
}

func (im InMemoryImpl) update(ID int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := im.DB.db[ID]; !ok {
		return ID, fmt.Errorf("%d key not found in DB", ID)
	}

	newID := shortURLToID(newURL.URL, chars)
	url := im.DB.db[ID]

	im.DB.db[newID] = url
	delete(im.DB.db, ID)

	return newID, nil
}

// save(user User) (primitive.ObjectID, error)
// findAll() ([]User, error)
func (mongoDAO MongoUserDaoImpl) save(newUser *User) (primitive.ObjectID, error) {
	res, err := mongoDAO.collection.InsertOne(mongoDAO.ctx, newUser)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	insertedID, _ := res.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

func (mongoDAO MongoUserDaoImpl) findAll() ([]User, error) {
	// A slice of tasks for storing the decoded documents
	var users []User

	cur, err := mongoDAO.collection.Find(mongoDAO.ctx, bson.D{})
	if err != nil {
		return users, err
	}

	for cur.Next(mongoDAO.ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return users, err
		}

		users = append(users, u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(mongoDAO.ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (mongoDAO MongoUserDaoImpl) filterUsers(filter interface{}) ([]*User, error) {
	// A slice of tasks for storing the decoded documents
	var users []*User

	cur, err := mongoDAO.collection.Find(mongoDAO.ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(mongoDAO.ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return users, err
		}

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(mongoDAO.ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (mongoDAO MongoUserDaoImpl) userExists(username string) (bool, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	users, err := mongoDAO.filterUsers(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return len(users) > 0, err
}

func (mongoDAO MongoUserDaoImpl) findByUsername(username string) (User, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	users, err := mongoDAO.filterUsers(filter)
	if err != nil {
		return User{}, fmt.Errorf("user '%s' not found in DB", username)
	}

	if len(users) == 0 {
		return User{}, fmt.Errorf("user '%s' not found in DB", username)
	}
	return (*users[0]), nil
}
