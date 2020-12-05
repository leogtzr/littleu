package main

func initializeRoutes() {
	router.Use(setUserStatus())

	// TODO: need to separate API vs regular links.
	// API:
	router.GET("/v", viewUrls)
	router.POST("/api/login", generateToken)

	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)
	router.POST("/u/shorturl", shorturl)
	router.POST("/u/changelink", changeLink)
	router.POST("/login", login)
	router.GET("/login", ensureNotLoggedIn(), showLoginPage)
	router.POST("/something", TokenAuthMiddleware(), CreateSomething)
	router.POST("/logout", TokenAuthMiddleware(), logout)
	router.GET("/register", ensureNotLoggedIn(), showRegistrationPage)
	router.POST("/register", ensureNotLoggedIn(), register)
}
