package main

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

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
