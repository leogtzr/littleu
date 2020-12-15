package main

func initializeRoutes() {
	router.Use(setUserStatus())

	router.POST("/api/login", generateToken)
	router.GET("/api/users", viewUsers)
	router.GET("/api/urls", viewURLs)

	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)
	router.POST("/u/shorturl", checkUserMiddleware(), shorturl)
	router.POST("/u/changelink", changeLink)
	router.POST("/login", login)
	router.GET("/login", ensureNotLoggedIn(), showLoginPage)
	router.POST("/logout", TokenAuthMiddleware(), logout)
	router.GET("/register", ensureNotLoggedIn(), showRegistrationPage)
	router.POST("/register", ensureNotLoggedIn(), register)
	router.GET("/session", checkSession)
}
