package main

import "github.com/spf13/viper"

func initializeRoutes(config *viper.Viper) {
	router.Use(setUserStatus())

	router.POST("/api/login", generateToken)
	router.GET("/api/users", viewUsers)
	router.GET("/api/urls", viewURLs)

	router.GET("/u/:url", urlStats(), redirectShortURL)
	router.GET("/", showIndexPage)
	router.POST("/u/shorturl", checkUserMiddleware(), shorturl)
	router.POST("/u/changelink", changeLink)
	router.POST("/login", login(config))
	router.GET("/login", ensureNotLoggedIn(), showLoginPage)
	router.POST("/logout", TokenAuthMiddleware(config), logout(config))
	router.GET("/register", ensureNotLoggedIn(), showRegistrationPage)
	router.POST("/register", ensureNotLoggedIn(), register(config))
	router.GET("/session", checkSession)

	// stats URLs
	router.GET("/stats", showStatsPage)
}
