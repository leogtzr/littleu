package main

import (
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func initializeRoutes() {
	router.Use(setUserStatus())

	// TODO: need to separate API vs regular links.
	router.GET("/v", viewUrls)
	router.GET("/us", viewUsers)
	router.POST("/api/login", generateToken)

	router.GET("/u/:url", redirectShortURL)
	router.GET("/", showIndexPage)
	router.POST("/u/shorturl", checkUserMiddleware(), shorturl)
	router.POST("/u/changelink", changeLink)
	router.POST("/login", login)
	router.GET("/login", ensureNotLoggedIn(), showLoginPage)
	router.POST("/something", TokenAuthMiddleware(), CreateSomething)
	router.POST("/logout", TokenAuthMiddleware(), logout)
	router.GET("/register", ensureNotLoggedIn(), showRegistrationPage)
	router.POST("/register", ensureNotLoggedIn(), register)
	router.GET("/session", checkSession)

	router.GET("/hello", func(c *gin.Context) {
		session := sessions.Default(c)

		if session.Get("hello") != "world" {
			session.Set("hello", User{
				User:      `el leix`,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Password:  `arroz`,
			})
			session.Save()
		}

		c.JSON(200, gin.H{"hello": session.Get("hello")})
	})

	router.GET("/other", func(c *gin.Context) {
		session := sessions.Default(c)

		c.JSON(200, gin.H{"hello": session.Get("hello")})
	})
}
