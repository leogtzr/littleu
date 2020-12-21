package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func showStatsPage(c *gin.Context) {
	session := sessions.Default(c)
	userFound := session.Get("user_logged_in")

	if userFound == nil {
		c.HTML(
			http.StatusInternalServerError,
			"error5xx.html",
			gin.H{
				"title":             "Error",
				"error_description": `You have to be logged in.`,
			},
		)

		return
	}

	// TODO: grab user's URLs...
	// TODO: pending.

	c.HTML(
		http.StatusOK,
		"stats.html",
		gin.H{
			"title": "Home",
		},
	)
}
