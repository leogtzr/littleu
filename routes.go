package main

func initializeRoutes() {
	router.GET("/u/:url", redirectShortURL)

	// Handle the index route
	router.GET("/", showIndexPage)

	// TODO: remove the following.
	router.GET("/article/view/:article_id", getArticle)

	router.POST("/u/shorturl", shorturl)
}
