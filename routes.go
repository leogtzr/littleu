package main

func initializeRoutes() {

	router.Use(setUserStatus())

	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)
	router.GET("/v", viewUrls)
	router.POST("/u/shorturl", shorturl)
	router.POST("/u/changelink", changeLink)
	router.POST("/login", login)
	router.POST("/something", TokenAuthMiddleware(), CreateSomething)
	router.POST("/logout", TokenAuthMiddleware(), logout)
	router.GET("/login", ensureNotLoggedIn(), showLoginPage)
	router.GET("/register", ensureNotLoggedIn(), showRegistrationPage)

}
