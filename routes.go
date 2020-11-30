package main

func initializeRoutes() {
	// Handle the index route
	router.GET("/", showIndexPage)

	// Handle GET requests at /article/view/some_article_id
	// TODO: remove the following.
	router.GET("/article/view/:article_id", getArticle)

	router.POST("/u/shorturl", shorturl)
}
