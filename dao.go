package main

import (
	"context"
	"encoding/binary"

	"database/sql"
	"errors"
	"fmt"
	"log"

	"crypto/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type randGenSrc struct{}

func (s *randGenSrc) Seed(seed int64) {}

func (s *randGenSrc) Uint64() (value uint64) {
	binary.Read(rand.Reader, binary.BigEndian, &value)
	return value
}

// URLDao ...
type URLDao interface {
	save(url URL, user *interface{}) (int, error)
	update(id int, oldURL, newURL URL) (int, error)
	findByID(id int) (URL, error)
	// findAll() (map[int]string, error)
}

// UserDAO ....
type UserDAO interface {
	addUser(username, password string) (interface{}, error)
	userExists(username string) (bool, error)
	findByUsername(username string) (interface{}, error)
	validateUserAndPassword(username, password string) (bool, error)
	findAll() ([]interface{}, error)
}

type memoryDB struct {
	db            map[int]string
	autoIncrement int
}

// InMemoryURLDAOImpl ...
type InMemoryURLDAOImpl struct {
	DB *memoryDB
}

// InMemoryUserDAOImpl ...
type InMemoryUserDAOImpl struct {
	db       map[string]UserInMemory
	rndIDGen randGenSrc
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

// PostgresqlUserImpl ...
type PostgresqlUserImpl struct {
	db *sql.DB
}

// PostgresqlURLDAOImpl ...
type PostgresqlURLDAOImpl struct {
	db *sql.DB
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

func (im InMemoryURLDAOImpl) save(url URL, user *interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	im.DB.autoIncrement++
	id := im.DB.autoIncrement
	im.DB.db[id] = url.URL

	return id, nil
}

func (im InMemoryURLDAOImpl) findAll() (map[int]string, error) {
	return im.DB.db, nil
}

func (im InMemoryURLDAOImpl) findByID(id int) (URL, error) {
	u, found := im.DB.db[id]
	if found {
		url := URL{
			URL: u,
		}

		return url, nil
	}

	return URL{}, fmt.Errorf("no url found for: %d", id)
}

func (im InMemoryURLDAOImpl) update(id int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := im.DB.db[id]; !ok {
		return id, fmt.Errorf("%d key not found in DB", id)
	}

	newID := shortURLToID(newURL.URL, chars)
	url := im.DB.db[id]

	im.DB.db[newID] = url
	delete(im.DB.db, id)

	return newID, nil
}

func (dao MongoUserDaoImpl) addUser(username, password string) (interface{}, error) {
	hashPassword := hashAndSalt([]byte(password))

	newUser := UserMongo{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      username,
		Password:  hashPassword,
	}

	_, err := dao.collection.InsertOne(dao.ctx, newUser)
	if err != nil {
		return UserMongo{}, err
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
		return users, err
	}

	for cur.Next(dao.ctx) {
		var u UserMongo

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
	cur.Close(dao.ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (dao MongoDBURLDAOImpl) filterURLs(filter interface{}) ([]URLDocument, error) {
	// A slice of tasks for storing the decoded documents
	var urls []URLDocument

	cur, err := dao.collection.Find(dao.ctx, filter)
	if err != nil {
		return urls, err
	}

	for cur.Next(dao.ctx) {
		var url URLDocument

		err := cur.Decode(&url)
		if err != nil {
			return urls, err
		}

		urls = append(urls, url)
	}

	if err := cur.Err(); err != nil {
		return urls, err
	}

	// once exhausted, close the cursor
	cur.Close(dao.ctx)

	if len(urls) == 0 {
		return urls, mongo.ErrNoDocuments
	}

	return urls, nil
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

func (dao MongoUserDaoImpl) findByUsername(username string) (interface{}, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	user, err := dao.filterUser(filter)
	if err != nil {
		return UserMongo{}, fmt.Errorf("user '%s' not found in DB", username)
	}

	return user, nil
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
		return -1, fmt.Errorf("error: incompatible types")
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
	if err != nil {
		return increment, err
	}

	return increment, nil
}

func (dao MongoDBURLDAOImpl) update(id int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	exists, err := dao.URLExists(id)
	if err != nil {
		return id, fmt.Errorf("error updating URL with %d id", id)
	}

	if !exists {
		return id, fmt.Errorf("%d key not found in DB", id)
	}

	newID := shortURLToID(newURL.URL, chars)

	exists, err = dao.URLExists(id)
	if err != nil {
		return id, fmt.Errorf("error updating URL with %d id", id)
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
		return -1, err
	}

	return newID, nil
}

func (dao MongoDBURLDAOImpl) findAll() (map[int]string, error) {
	filter := bson.D{}

	allURLs, err := dao.filterURLs(filter)
	if err != nil {
		return map[int]string{}, nil
	}

	urlMap := map[int]string{}

	for _, u := range allURLs {
		urlMap[u.ShortID] = u.URL
	}

	return urlMap, nil
}

func (dao MongoDBURLDAOImpl) findByID(id int) (URL, error) {
	filter := bson.D{
		primitive.E{Key: "shortid", Value: id},
	}

	var urlDoc URLDocument

	err := dao.collection.FindOne(dao.ctx, filter).Decode(&urlDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return URL{}, err
		}
	}

	url := URL{}
	url.URL = urlDoc.URL

	return URL{}, nil
}

func (dao MongoDBURLDAOImpl) getMaxShortID() (int, error) {
	var url URLDocument

	options := options.FindOne()
	options.SetSort(bson.D{{"shortid", -1}})

	err := dao.collection.FindOne(dao.ctx, bson.D{}, options).Decode(&url)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}

		return -1, err
	}

	return url.ShortID, nil
}

func (dao PostgresqlUserImpl) addUser(username, password string) (interface{}, error) {
	hashPassword := hashAndSalt([]byte(password))

	user := UserPostgresql{
		ID:        -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      username,
		Password:  hashPassword,
	}

	createUserSQL := `
		INSERT INTO users (created_at, updated_at, username, password) VALUES ($1, $2, $3, $4) RETURNING id
	`

	_, err := dao.db.Exec(createUserSQL, user.CreatedAt, user.UpdatedAt, user.User, user.Password)
	if err != nil {
		return -1, err
	}

	return user, nil
}

func (dao PostgresqlUserImpl) findByUsername(username string) (interface{}, error) {
	var user UserPostgresql

	query := `select id, username, password, created_at, updated_at from users where username = $1`

	err :=
		dao.db.QueryRow(query, username).Scan(&user.ID, &user.User, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserPostgresql{}, fmt.Errorf("user '%s' not found in DB", username)
		}

		return UserPostgresql{}, err
	}

	if user.User == username {
		return user, nil
	}

	return UserPostgresql{}, fmt.Errorf("user '%s' not found in DB", username)
}

func (dao PostgresqlUserImpl) userExists(username string) (bool, error) {
	var user UserPostgresql

	query := `select username from users where username = $1`

	err := dao.db.QueryRow(query, username).Scan(&user.User)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	if user.User == username {
		return true, nil
	}

	return false, nil
}

func (dao PostgresqlURLDAOImpl) getMaxShortID() (int, error) {
	var id int

	query := `SELECT coalesce(max(short_id), 0) FROM urls`

	err := dao.db.QueryRow(query).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 1, nil
		}

		return -1, err
	}

	return id, nil
}

func (dao PostgresqlURLDAOImpl) save(url URL, user *interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	maxID, err := dao.getMaxShortID()
	if err != nil {
		return -1, err
	}

	maxID++

	u, ok := (*user).(*UserPostgresql)
	if !ok {
		return -1, fmt.Errorf("error: incompatible types")
	}

	createURLSQL := `
		INSERT INTO urls (created_at, updated_at, url, short_id, user_id) values($1, $2, $3, $4, $5) RETURNING id
	`

	_, err = dao.db.Exec(createURLSQL, time.Now(), time.Now(), url.URL, maxID, u.ID)
	if err != nil {
		return -1, err
	}

	return maxID, nil
}

func (dao PostgresqlURLDAOImpl) update(id int, oldURL, newURL URL) (int, error) {
	return -1, nil
}

// func (dao PostgresqlURLDAOImpl) findAll() (map[int]string, error) {
// 	return map[int]string{}, nil
// }

func (dao PostgresqlURLDAOImpl) findByID(id int) (URL, error) {
	return URL{}, nil
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
		return false, fmt.Errorf("error: incompatible types")
	}

	hashFromDatabase := []byte(u.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (dao PostgresqlUserImpl) validateUserAndPassword(username, password string) (bool, error) {
	user, err := dao.findByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	u, ok := user.(UserPostgresql)
	if !ok {
		return false, fmt.Errorf("error: incompatible types")
	}

	hashFromDatabase := []byte(u.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (dao InMemoryUserDAOImpl) addUser(username, password string) (interface{}, error) {
	hashPassword := hashAndSalt([]byte(password))

	id := dao.rndIDGen.Uint64()

	newUser := UserInMemory{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      username,
		Password:  hashPassword,
	}

	dao.db[username] = newUser

	return newUser, nil
}

func (dao InMemoryUserDAOImpl) userExists(username string) (bool, error) {
	_, exists := dao.db[username]
	if !exists {
		return false, nil
	}

	return true, nil
}

func (dao InMemoryUserDAOImpl) findByUsername(username string) (interface{}, error) {
	user, exists := dao.db[username]
	if !exists {
		return UserInMemory{}, fmt.Errorf("user '%s' not found in DB", username)
	}

	return user, nil
}

func (dao InMemoryUserDAOImpl) validateUserAndPassword(username, password string) (bool, error) {
	user, err := dao.findByUsername(username)
	if err != nil {
		return false, err
	}

	u, ok := user.(UserInMemory)
	if !ok {
		return false, fmt.Errorf("error: incompatible types")
	}

	hashFromDatabase := []byte(u.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (dao InMemoryUserDAOImpl) findAll() ([]interface{}, error) {

	users := []interface{}{}

	for _, v := range dao.db {
		users = append(users, v)
	}

	return users, nil
}

func (dao MongoUserDaoImpl) findAll() ([]interface{}, error) {
	filter := bson.D{}

	var us []interface{}

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

func (dao PostgresqlUserImpl) findAll() ([]interface{}, error) {

	query := `SELECT id, username, password, created_at, updated_at FROM users`

	var us []interface{}

	rows, err := dao.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var user UserPostgresql
		if err :=
			rows.Scan(&user.ID, &user.User, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return []interface{}{}, err
		}
		us = append(us, user)
	}
	if err := rows.Err(); err != nil {
		return []interface{}{}, err
	}

	return us, nil
}
