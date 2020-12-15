package main

import (
	"errors"
	"fmt"
)

var (
	errNOURLFound        = errors.New("no url found")
	errUserNotFound      = errors.New("user not found")
	errIncompatibleTypes = errors.New("incompatible types")
)

func errorURLNotFound(url int) error {
	return fmt.Errorf("errorURLNotFound %w : %d id", errNOURLFound, url)
}

func errorUserNotFound(username string) error {
	return fmt.Errorf("errorUserNotFound %w : %s", errUserNotFound, username)
}

func errorIncompatibleTypes() error {
	return fmt.Errorf("errorIncompatibleTypes %w", errIncompatibleTypes)
}
