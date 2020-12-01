package main

func initializeRoutes() {
	// Handle the index route
	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)

	// TODO: remove the following.
	router.GET("/article/view/:article_id", getArticle)

	router.POST("/u/shorturl", shorturl)
}
