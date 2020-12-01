package main

import (
	"log"
)

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

func (im InMemoryImpl) save(u URL) error {
	mu.Lock()

	im.DB.autoIncrement++
	id := im.DB.autoIncrement
	im.DB.db[id] = u.URL

	mu.Unlock()
	return nil
}

func (im InMemoryImpl) findAll() ([]URL, error) {
	// TODO: implement this.
	var urls []URL
	return urls, nil

}
