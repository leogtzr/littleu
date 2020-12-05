package main

import (
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

func idToShortURL(id int, mChars []rune) string {
	shortURL := ""
	mapCharsSize := len(mChars)

	for id > 0 {
		shortURL += string(mChars[id%mapCharsSize])
		id /= mapCharsSize
	}

	return reverse(shortURL)
}

func shortURLToID(shortURL string, mChars []rune) int {
	mapCharsSize := len(mChars)
	id := 0
	for _, i := range shortURL {
		c := int(i)
		if c >= int('a') && c <= int('z') {
			id = id*mapCharsSize + c - int('a')
		} else if c >= int('A') && c <= int('Z') {
			id = id*mapCharsSize + c - int('Z') + 26
		} else {
			id = id*mapCharsSize + c - int('0') + 52
		}
	}
	return id
}

func readConfig(filename, configPath string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(configPath)
	v.SetConfigType("env")
	err := v.ReadInConfig()
	return v, err
}

// 	id := 12345
// 	shortURL := idToShortURL(id, chars)
//  url := shortURLToID(shortURL, chars))
// collection = client.Database("littleu").Collection("user")
func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) (bool, error) { // Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TODO: move this to another file:
// TODO: DAO...
func createUser(newUser *NewUser) error {
	_, err := collection.InsertOne(ctx, newUser)
	return err
}

// NewUser ...
func filterUsers(filter interface{}) ([]*NewUser, error) {
	// A slice of tasks for storing the decoded documents
	var users []*NewUser

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(ctx) {
		var u NewUser
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
	cur.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func userExists(username string) (bool, error) {

	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	users, err := filterUsers(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return len(users) > 0, err
}

func validateUserAndPassword(username, password string) (bool, error) {

	filter := bson.D{
		primitive.E{Key: "user", Value: username},
	}

	users, err := filterUsers(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	if len(users) == 0 {
		return false, nil
	}

	userFromDatabase := users[0]
	hashFromDatabase := []byte(userFromDatabase.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func validateNewUserFields(user, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("The password can't be empty")
	}
	if strings.TrimSpace(user) == "" {
		return errors.New("The username can't be empty")
	}
	return nil
}
