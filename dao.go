package main

import (
	"fmt"
	"log"
)

// DBHandler ...
// TODO: fix the following interface.
type DBHandler interface {
	save(u URL) (int, error)
	findAll() ([]URL, error)
	findByID(id int) (URL, error)
}

type memoryDB struct {
	db            map[int]string
	autoIncrement int
}

// InMemoryImpl ...
type InMemoryImpl struct {
	DB *memoryDB
}

func factory(engine string) *DBHandler {
	var dao DBHandler
	switch engine {
	case "memory":
		dao = InMemoryImpl{
			DB: &memoryDB{
				db: map[int]string{},
			},
		}
	default:
		log.Fatalf("error: wrong engine: %s", engine)
		return nil
	}
	return &dao
}

func (im InMemoryImpl) save(u URL) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	im.DB.autoIncrement++
	id := im.DB.autoIncrement
	im.DB.db[id] = u.URL

	return id, nil
}

func (im InMemoryImpl) findAll() ([]URL, error) {
	// TODO: implement this.
	var urls []URL
	return urls, nil
}

func (im InMemoryImpl) findByID(id int) (URL, error) {
	u, found := im.DB.db[id]
	if found {
		url := URL{
			URL: u,
		}
		return url, nil
	}
	return URL{}, fmt.Errorf("no url found for: %d", id)

}
