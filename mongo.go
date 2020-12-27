package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

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

// StatsPostgresqlImpl ...
type StatsMongoImpl struct {
	collection *mongo.Collection
	ctx        context.Context
}

// URLExists ...
func (dao MongoDBURLDAOImpl) URLExists(urlID int) (bool, error) {
	filter := bson.D{
		primitive.E{Key: "shortid", Value: urlID},
	}

	urls, err := dao.filterURLs(filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return len(urls) > 0, err
}

func (dao MongoDBURLDAOImpl) save(url URL, user *interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	increment := 0

	maxURLID, err := dao.getMaxShortID()
	if err != nil {
		return -1, err
	}

	increment = maxURLID
	increment++

	u, ok := (*user).(*UserMongo)
	if !ok {
		return -1, errorIncompatibleTypes()
	}

	urlDoc := URLDocument{
		ShortID:   increment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        primitive.NewObjectID(),
		URL:       url.URL,
		UserID:    u.ID,
	}

	_, err = dao.collection.InsertOne(dao.ctx, urlDoc)

	return increment, fmt.Errorf("error inserting url: %w", err)
}

func (dao MongoDBURLDAOImpl) filterURLs(filter interface{}) ([]URLDocument, error) {
	// A slice of tasks for storing the decoded documents
	var urls []URLDocument

	cur, err := dao.collection.Find(dao.ctx, filter)
	if err != nil {
		return urls, fmt.Errorf("error finding user: %w", err)
	}

	for cur.Next(dao.ctx) {
		var url URLDocument

		err := cur.Decode(&url)
		if err != nil {
			return urls, fmt.Errorf("error converting user: %w", err)
		}

		urls = append(urls, url)
	}

	if err := cur.Err(); err != nil {
		return urls, fmt.Errorf("error closing db cursor: %v", err)
	}

	// once exhausted, close the cursor
	_ = cur.Close(dao.ctx)

	if len(urls) == 0 {
		return urls, mongo.ErrNoDocuments
	}

	return urls, nil
}

func (dao MongoDBURLDAOImpl) findByID(id int) (URL, error) {
	filter := bson.D{
		primitive.E{Key: "shortid", Value: id},
	}

	var urlDoc URLDocument

	err := dao.collection.FindOne(dao.ctx, filter).Decode(&urlDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return URL{}, fmt.Errorf("error, no documents: %v", err)
		}
	}

	url := URL{}
	url.URL = urlDoc.URL

	return url, nil
}

func (dao MongoDBURLDAOImpl) getMaxShortID() (int, error) {
	var url URLDocument

	sortOptions := options.FindOne()
	sortOptions.SetSort(bson.D{{"shortid", -1}})

	err := dao.collection.FindOne(dao.ctx, bson.D{}, sortOptions).Decode(&url)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}

		return -1, fmt.Errorf("error no documents: %v", err)
	}

	return url.ShortID, nil
}

func (dao MongoDBURLDAOImpl) update(id int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	exists, err := dao.URLExists(id)
	if err != nil {
		return id, errorUpdatingURL(id)
	}

	if !exists {
		return id, errorKeyNotFoundInDB(id)
	}

	newID := shortURLToID(newURL.URL, chars)

	exists, err = dao.URLExists(id)
	if err != nil {
		return id, errorUpdatingURL(id)
	}

	if !exists {
		return id, fmt.Errorf("URL %s already exists, pick a different one", newURL.URL)
	}

	_, err = dao.collection.UpdateOne(
		ctx,
		primitive.E{Key: "shortid", Value: id},
		bson.D{
			{"$set", bson.D{{"shortid", newID}}},
		},
	)

	if err != nil {
		return -1, fmt.Errorf("error updating url: %v", err)
	}

	return newID, nil
}

func (dao MongoDBURLDAOImpl) findAllByUser(user *interface{}) ([]URLStat, error) {
	userDB, ok := (*user).(*UserMongo)
	if !ok {
		return []URLStat{}, errorIncompatibleTypes()
	}

	filter := bson.D{
		primitive.E{Key: "user_id", Value: userDB.ID},
	}

	allURLs, err := dao.filterURLs(filter)

	if err != nil {
		return []URLStat{}, nil
	}

	return toURLStat(&allURLs), nil
}

func toURLStat(urlDocs *[]URLDocument) []URLStat {
	urls := []URLStat{}

	for _, u := range *urlDocs {
		urls = append(urls, URLStat{
			ShortID: u.ShortID,
			Url:     u.URL,
		})
	}

	return urls
}

func (dao MongoUserDaoImpl) addUser(username, password string) (interface{}, error) {
	hashPassword := password

	newUser := UserMongo{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      username,
		Password:  hashPassword,
	}

	_, err := dao.collection.InsertOne(dao.ctx, newUser)
	if err != nil {
		return UserMongo{}, fmt.Errorf("error inserting user: %w", err)
	}

	return newUser, nil
}

func (dao MongoUserDaoImpl) filterUser(filter interface{}) (UserMongo, error) {
	var user UserMongo
	err := dao.collection.FindOne(dao.ctx, filter).Decode(&user)

	return user, err
}

func (dao MongoUserDaoImpl) filterUsers(filter interface{}) ([]UserMongo, error) {
	var users []UserMongo

	cur, err := dao.collection.Find(dao.ctx, filter)
	if err != nil {
		return users, fmt.Errorf("error getting user: %v", err)
	}

	for cur.Next(dao.ctx) {
		var u UserMongo

		err := cur.Decode(&u)
		if err != nil {
			return users, fmt.Errorf("error converting user: %v", err)
		}

		users = append(users, u)
	}

	if err := cur.Err(); err != nil {
		return users, fmt.Errorf("error: %v", err)
	}

	// once exhausted, close the cursor
	_ = cur.Close(dao.ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (dao MongoUserDaoImpl) userExists(username string) (bool, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	_, err := dao.filterUser(filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (dao MongoUserDaoImpl) findByUsername(username string) (interface{}, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	user, err := dao.filterUser(filter)
	if err != nil {
		return UserMongo{}, errorUserNotFound(username)
	}

	return user, nil
}

func (dao MongoUserDaoImpl) validateUserAndPassword(username, password string) (bool, error) {
	user, err := dao.findByUsername(username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		return false, err
	}

	u, ok := user.(UserMongo)
	if !ok {
		return false, errorIncompatibleTypes()
	}

	hashFromDatabase := []byte(u.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (dao MongoUserDaoImpl) findAll() ([]interface{}, error) {
	filter := bson.D{}

	us := []interface{}{}

	users, err := dao.filterUsers(filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []interface{}{}, nil
		}
	}

	for _, u := range users {
		us = append(us, u)
	}

	return us, nil
}

func (dao StatsMongoImpl) save(URL string, headers *map[string][]string, user *interface{}) (int, error) {
	return -1, nil
}

func (dao StatsMongoImpl) findByShortID(shortID int) ([]interface{}, error) {
	return []interface{}{}, nil
}
