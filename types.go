package main

// URL ...
type URL struct {
	URL string `form:"url"`
}

// DBHandler ...
// TODO: fix the following interface.
type DBHandler interface {
	save(u URL) error
	findAll() ([]URL, error)
}

type memoryDB struct {
	db            map[int]string
	autoIncrement int
}
