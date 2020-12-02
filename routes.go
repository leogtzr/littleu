package main

func initializeRoutes() {
	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)
	router.GET("/v", viewUrls)
	router.POST("/u/shorturl", shorturl)
	router.POST("/u/changelink", changeLink)
}
