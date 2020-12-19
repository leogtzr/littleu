package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PostgresqlUserImpl ...
type PostgresqlUserImpl struct {
	db *sql.DB
}

// PostgresqlURLDAOImpl ...
type PostgresqlURLDAOImpl struct {
	db *sql.DB
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
		return -1, fmt.Errorf("error adding user: %v", err)
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
			return UserPostgresql{}, errorUserNotFound(username)
		}

		return UserPostgresql{}, fmt.Errorf("error getting username: %v", err)
	}

	if user.User == username {
		return user, nil
	}

	return UserPostgresql{}, errorUserNotFound(username)
}

func (dao PostgresqlUserImpl) userExists(username string) (bool, error) {
	var user UserPostgresql

	query := `select username from users where username = $1`

	err := dao.db.QueryRow(query, username).Scan(&user.User)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("error getting user: %v", err)
	}

	if user.User == username {
		return true, nil
	}

	return false, nil
}

// URLExists ...
func (dao PostgresqlURLDAOImpl) URLExists(urlID int) (bool, error) {
	query := `select short_id from urls where short_id = $1`

	var shortID int

	err := dao.db.QueryRow(query, urlID).Scan(&shortID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("error getting url: %v", err)
	}

	if shortID == urlID {
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

		return -1, fmt.Errorf("error getting max url id: %v", err)
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
		return -1, errorIncompatibleTypes()
	}

	createURLSQL := `
		INSERT INTO urls (created_at, updated_at, url, short_id, user_id) values($1, $2, $3, $4, $5) RETURNING id
	`

	_, err = dao.db.Exec(createURLSQL, time.Now(), time.Now(), url.URL, maxID, u.ID)
	if err != nil {
		return -1, fmt.Errorf("error creating url: %v", err)
	}

	return maxID, nil
}

func (dao PostgresqlURLDAOImpl) update(id int, oldURL, newURL URL) (int, error) {
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

	stmtQuery := `update urls set short_id = $1 where short_id = $2`

	_, err = dao.db.Exec(stmtQuery, newID, id)
	if err != nil {
		return -1, fmt.Errorf("error updating url: %v", err)
	}

	return newID, nil
}

func (dao PostgresqlURLDAOImpl) findAll() (map[int]string, error) {
	query := `SELECT short_id, url FROM urls`

	urls := map[int]string{}

	rows, err := dao.db.Query(query)
	if err != nil {
		return map[int]string{}, fmt.Errorf("error getting urls: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int

		var url string

		if err := rows.Scan(&id, &url); err != nil {
			return map[int]string{}, fmt.Errorf("error getting urls: %v", err)
		}

		urls[id] = url
	}

	if err := rows.Err(); err != nil {
		return map[int]string{}, fmt.Errorf("error closing cursor: %v", err)
	}

	return urls, nil
}

func (dao PostgresqlURLDAOImpl) findByID(id int) (URL, error) {
	query := `select url from urls where short_id = $1`
	url := URL{}

	err := dao.db.QueryRow(query, id).Scan(&url.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URL{}, errorURLNotFound(id)
		}

		return URL{}, fmt.Errorf("error getting url: %v", err)
	}

	return url, nil
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
		return false, errorIncompatibleTypes()
	}

	hashFromDatabase := []byte(u.Password)
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (dao PostgresqlUserImpl) findAll() ([]interface{}, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users`

	var us []interface{}

	rows, err := dao.db.Query(query)
	if err != nil {
		return []interface{}{}, fmt.Errorf("error getting users: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user UserPostgresql
		if err :=
			rows.Scan(&user.ID, &user.User, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return []interface{}{}, fmt.Errorf("error getting users: %v", err)
		}

		us = append(us, user)
	}

	if err := rows.Err(); err != nil {
		return []interface{}{}, fmt.Errorf("error getting users: %v", err)
	}

	return us, nil
}
