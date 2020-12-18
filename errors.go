package main

import (
	"errors"
	"fmt"
)

var (
	errNOURLFound         = errors.New("no url found")
	errUserNotFound       = errors.New("user not found")
	errIncompatibleTypes  = errors.New("incompatible types")
	errPasswordFieldEmpty = errors.New("password cannot be empty")
	errUsernameFieldEmpty = errors.New("username cannot be empty")
	errKeyNotFoundInDB    = errors.New("key not found")
	errUpdatingURL        = errors.New("updating url")
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

func errorPasswordFieldEmpty() error {
	return fmt.Errorf("errPasswordFieldEmpty %w", errPasswordFieldEmpty)
}

func errorUsernameFieldEmpty() error {
	return fmt.Errorf("errUsernameFieldEmpty %w", errUsernameFieldEmpty)
}

func errorKeyNotFoundInDB(id int) error {
	return fmt.Errorf("errKeyNotFoundInDB %w : %d id", errKeyNotFoundInDB, id)
}

func errorUpdatingURL(id int) error {
	return fmt.Errorf("errUpdatingURL %w : %d id", errUpdatingURL, id)
}
