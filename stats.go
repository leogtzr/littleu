package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func showStatsPage(c *gin.Context) {
	// c.JSON(http.StatusOK, gin.H{"message": "OK"})
	c.HTML(
		http.StatusOK,
		"stats.html",
		gin.H{
			"title": "Home",
		},
	)
}
