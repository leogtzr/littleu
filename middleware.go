package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// This middleware ensures that a request will be aborted with an error
// if the user is already logged in.
func ensureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If there's no error or if the token is not empty
		// the user is already logged in
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)

		if loggedIn {
			// if token, err := c.Cookie("token"); err == nil || token != "" {

			/*
				Here we need to decide between two things.
				If the user is already logged in and tries to hit the /register url,
					do you show him/her an error StatusUnauthorized(401) or
					do you send him/her to the index.html page?
			*/

			// c.AbortWithStatus(http.StatusUnauthorized)
			c.HTML(
				http.StatusOK,
				"index.html",
				gin.H{
					"title": "Home",
				},
			)
		}
	}
}
