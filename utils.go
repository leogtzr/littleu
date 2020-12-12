package main

import (
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/spf13/viper"
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

func validateNewUserFields(user, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("The password can't be empty")
	}
	if strings.TrimSpace(user) == "" {
		return errors.New("The username can't be empty")
	}
	return nil
}
