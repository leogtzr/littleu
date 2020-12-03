package main

// URL ...
type URL struct {
	URL string `form:"url"`
}

// URLChange ...
type URLChange struct {
	ShortURL string `form:"url"`
	NewURL   string `form:"new_url"`
}

// User ...
type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
