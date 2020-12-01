package main

import (
	"unicode/utf8"

	"github.com/spf13/viper"
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
