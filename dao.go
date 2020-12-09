package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// "github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "github.com/lib/pq"
)

// URLDao ...
type URLDao interface {
	save(url URL, user *interface{}) (int, error)
	update(ID int, oldURL, newURL URL) (int, error)
	findAll() (map[int]string, error)
	findByID(ID int) (URL, error)
}

// UserDAO ....
type UserDAO interface {
	save(user *User) (interface{}, error)
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
		dao = InMemoryImpl{
			DB: &memoryDB{
				db: map[int]string{},
			},
		}
	case "mongo":
		var err error

		mongoClientOptions = options.Client().ApplyURI(envConfig.GetString("MONGO_URI"))
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
		dsn := envConfig.GetString("POSTGRES_DSN")
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
	// TODO: missing "memory" here
	switch engine {
	case "mongo":

		var collection *mongo.Collection
		collection = mongoClient.Database("littleu").Collection("user")

		userDAO = MongoUserDaoImpl{
			collection: collection,
			ctx:        ctx,
		}

	case "postgresql":

		dsn := envConfig.GetString("POSTGRES_DSN")
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

func (im InMemoryImpl) save(url URL, user *interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	im.DB.autoIncrement++
	id := im.DB.autoIncrement
	im.DB.db[id] = url.URL

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
func (dao MongoUserDaoImpl) save(newUser *User) (interface{}, error) {
	res, err := dao.collection.InsertOne(dao.ctx, newUser)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	insertedID, _ := res.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

func (dao MongoUserDaoImpl) findAll() ([]User, error) {
	// A slice of tasks for storing the decoded documents
	var users []User

	cur, err := dao.collection.Find(dao.ctx, bson.D{})
	if err != nil {
		return users, err
	}

	for cur.Next(dao.ctx) {
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
	cur.Close(dao.ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func (dao MongoUserDaoImpl) filterUser(filter interface{}) (User, error) {
	var user User
	err := dao.collection.FindOne(dao.ctx, filter).Decode(&user)
	return user, err
}

func (dao MongoUserDaoImpl) filterUsers(filter interface{}) ([]User, error) {
	var users []User

	cur, err := dao.collection.Find(dao.ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(dao.ctx) {
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
		if err == mongo.ErrNoDocuments {
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
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return len(urls) > 0, err
}

func (dao MongoUserDaoImpl) findByUsername(username string) (User, error) {
	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	user, err := dao.filterUser(filter)
	if err != nil {
		return User{}, fmt.Errorf("user '%s' not found in DB", username)
	}
	return user, nil
}

func (dao MongoDBURLDAOImpl) save(url URL, user *interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	increment := 0

	maxURLID, err := dao.getMaxShortID()
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
		URL:       url.URL,
	}

	_, err = dao.collection.InsertOne(dao.ctx, urlDoc)
	if err != nil {
		return increment, err
	}
	return increment, nil
}

// TODO: finish impl for this.
func (dao MongoDBURLDAOImpl) update(ID int, oldURL, newURL URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	exists, err := dao.URLExists(ID)
	if err != nil {
		return ID, fmt.Errorf("error updating URL with %d id", ID)
	}
	if !exists {
		return ID, fmt.Errorf("%d key not found in DB", ID)
	}

	newID := shortURLToID(newURL.URL, chars)
	exists, err = dao.URLExists(ID)
	if err != nil {
		return ID, fmt.Errorf("URL %s already exists, pick a different one", newURL.URL)
	}

	fmt.Printf("debug - newID is: %d\n", newID)

	_, err = dao.collection.UpdateOne(
		ctx,
		primitive.E{Key: "shortid", Value: ID},
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

func (dao MongoDBURLDAOImpl) findByID(ID int) (URL, error) {

	filter := bson.D{
		primitive.E{Key: "shortid", Value: ID},
	}

	var urlDoc URLDocument
	err := dao.collection.FindOne(dao.ctx, filter).Decode(&urlDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return URL{}, err
		}
	}

	url := URL{}
	url.URL = urlDoc.URL

	return URL{}, nil
}

func (dao MongoDBURLDAOImpl) getMaxShortID() (URLDocument, error) {
	var url URLDocument

	options := options.FindOne()
	options.SetSort(bson.D{{"shortid", -1}})

	err := dao.collection.FindOne(dao.ctx, bson.D{}, options).Decode(&url)
	if err != nil {
		return URLDocument{}, err
	}

	return url, nil
}

func (dao PostgresqlUserImpl) save(user *User) (interface{}, error) {
	createUserSQL := `
		INSERT INTO users (created_at, updated_at, username, password) VALUES ($1, $2, $3, $4) RETURNING id
	`

	lastID, err := dao.db.Exec(createUserSQL, user.CreatedAt, user.UpdatedAt, user.User, user.Password)

	if err != nil {
		return -1, err
	}

	return lastID, nil
}

func (dao PostgresqlUserImpl) findAll() ([]User, error) {

	var users []User
	query := `select username, password, created_at, updated_at from users`

	rows, err := dao.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("debug 1")
			return []User{}, nil
		}
		fmt.Printf("debug 2")
		return []User{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.User, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
			fmt.Printf("debug 3")
			return []User{}, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return []User{}, err
	}

	return users, nil

}

func (dao PostgresqlUserImpl) findByUsername(username string) (User, error) {

	var user User
	query := `select username, password, created_at, updated_at from users where username = $1`

	err := dao.db.QueryRow(query, username).Scan(&user.User, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user '%s' not found in DB", username)
		}
		return User{}, err
	}

	if user.User == username {
		return user, nil
	}

	return User{}, fmt.Errorf("user '%s' not found in DB", username)
}

func (dao PostgresqlUserImpl) userExists(username string) (bool, error) {
	var user User
	query := `select username from users where username = $1`

	err := dao.db.QueryRow(query, username).Scan(&user.User)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if user.User == username {
		return true, nil
	}

	return false, nil
}

func (dao PostgresqlURLDAOImpl) save(url URL, user *interface{}) (int, error) {
	return -1, nil
}

func (dao PostgresqlURLDAOImpl) update(ID int, oldURL, newURL URL) (int, error) {
	return -1, nil
}

func (dao PostgresqlURLDAOImpl) findAll() (map[int]string, error) {
	return map[int]string{}, nil
}

func (dao PostgresqlURLDAOImpl) findByID(ID int) (URL, error) {
	return URL{}, nil
}
