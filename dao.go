package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// "github.com/mitchellh/mapstructure"

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

// MongoDBURLDAOImpl ...
type MongoDBURLDAOImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

// MongoUserDaoImpl ...
type MongoUserDaoImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

func factoryURLDao(engine string, config *viper.Viper) *URLDao {
	var dao URLDao
	switch engine {
	case "memory":
		dao = InMemoryImpl{
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

		var collection *mongo.Collection
		collection = mongoClient.Database("littleu").Collection("user")

		userDAO = MongoUserDaoImpl{
			collection: collection,
			ctx:        ctx,
		}

	default:
		log.Fatalf("error: wrong engine: %s", engine)
		return nil
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

func (mongoDAO MongoUserDaoImpl) filterUser(filter interface{}) (User, error) {
	var user User
	err := mongoDAO.collection.FindOne(mongoDAO.ctx, filter).Decode(&user)
	return user, err
}

func (mongoDAO MongoUserDaoImpl) filterUsers(filter interface{}) ([]User, error) {
	var users []User

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

func (mongoDAO MongoDBURLDAOImpl) filterURLs(filter interface{}) ([]*interface{}, error) {
	// A slice of tasks for storing the decoded documents
	var urls []*interface{}

	cur, err := mongoDAO.collection.Find(mongoDAO.ctx, filter)
	if err != nil {
		return urls, err
	}

	for cur.Next(mongoDAO.ctx) {
		var url interface{}
		err := cur.Decode(&url)
		if err != nil {
			return urls, err
		}

		urls = append(urls, &url)
	}

	if err := cur.Err(); err != nil {
		return urls, err
	}

	// once exhausted, close the cursor
	cur.Close(mongoDAO.ctx)

	if len(urls) == 0 {
		return urls, mongo.ErrNoDocuments
	}

	return urls, nil
}

func (mongoDAO MongoUserDaoImpl) userExists(username string) (bool, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	_, err := mongoDAO.filterUser(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// URLExists ...
func (mongoDAO MongoDBURLDAOImpl) URLExists(urlID int) (bool, error) {
	filter := bson.D{
		primitive.E{Key: "shortid", Value: urlID},
	}

	urls, err := mongoDAO.filterURLs(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return len(urls) > 0, err
}

func (mongoDAO MongoUserDaoImpl) findByUsername(username string) (User, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	user, err := mongoDAO.filterUser(filter)
	if err != nil {
		return User{}, fmt.Errorf("user '%s' not found in DB", username)
	}
	return user, nil
}

func (mongoURLDAO MongoDBURLDAOImpl) save(u URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	increment := 0

	maxURLID, err := mongoURLDAO.getMaxShortID()
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return -1, err
		}
	}

	increment = maxURLID.ShortID
	increment++

	// Convert u(URL) to URLDocument
	urlDoc := URLDocument{
		ShortID:   increment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        primitive.NewObjectID(),
		URL:       u.URL,
	}

	_, err = mongoURLDAO.collection.InsertOne(mongoURLDAO.ctx, urlDoc)
	if err != nil {
		return increment, err
	}
	return increment, nil
}

// TODO: finish impl for this.
func (mongoURLDAO MongoDBURLDAOImpl) update(ID int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	exists, err := mongoURLDAO.URLExists(ID)
	if err != nil {
		return ID, fmt.Errorf("error updating URL with %d id", ID)
	}
	if !exists {
		return ID, fmt.Errorf("%d key not found in DB", ID)
	}

	newID := shortURLToID(newURL.URL, chars)
	exists, err = mongoURLDAO.URLExists(ID)
	if err != nil {
		return ID, fmt.Errorf("URL %s already exists, pick a different one", newURL.URL)
	}

	fmt.Printf("debug - newID is: %d\n", newID)

	_, err = mongoURLDAO.collection.UpdateOne(
		ctx,
		bson.M{"shortid": ID},
		bson.D{
			{"$set", bson.D{{"shortid", newID}}},
		},
	)

	if err != nil {
		return -1, err
	}

	return newID, nil
}

// TODO: finish impl for this:
func (mongoURLDAO MongoDBURLDAOImpl) findAll() (map[int]string, error) {
	return map[int]string{}, nil
}

// TODO: finish impl for this.
func (mongoURLDAO MongoDBURLDAOImpl) findByID(ID int) (URL, error) {
	return URL{}, nil
}

func (mongoURLDAO MongoDBURLDAOImpl) getMaxShortID() (URLDocument, error) {
	var url URLDocument

	options := options.FindOne()
	options.SetSort(bson.D{{"shortid", -1}})

	err := mongoURLDAO.collection.FindOne(mongoURLDAO.ctx, bson.D{}, options).Decode(&url)
	if err != nil {
		return URLDocument{}, err
	}

	return url, nil
}
