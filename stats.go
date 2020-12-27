package main

import (
	"fmt"
	"github.com/Showmax/go-fqdn"
	"github.com/spf13/viper"
	"net"

	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func showStatsPage(config *viper.Viper) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		urlStats, err := (*urlDAO).findAllByUser(&userFound)
		if err != nil {
			c.HTML(
				http.StatusInternalServerError,
				"error5xx.html",
				gin.H{
					"title":             "Error",
					"error_description": err.Error(),
				},
			)

			return
		}

		fqdnHostName, err := fqdn.FqdnHostname()
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}

		domain := net.JoinHostPort(fqdnHostName, config.GetString("port"))

		urlsFull := urlsToFullStat(&urlStats)

		c.HTML(
			http.StatusOK,
			"stats.html",
			gin.H{
				"title": "URL Stats",
				"domain": domain,
				"urls":  urlsFull,
			},
		)
	}
}

func urlStats() gin.HandlerFunc {
	return func(c *gin.Context) {

		shortURLParam := c.Param("url")
		fmt.Printf("The url to redirect to is: [%s]\n", shortURLParam)

		session := sessions.Default(c)
		user := session.Get("user_logged_in")

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "unauthorized"})
			c.Abort()
		}

		// @TODO: remove the following code.
		fmt.Println("debug headers - begin")

		for headerName, headers := range c.Request.Header {
			fmt.Printf("header -> [%s]\n", headerName)
			for _, header := range headers {
				fmt.Printf("\t[%s]\n", header)
			}
		}

		fmt.Println("debug headers - end")
	}
}

func viewStats(c *gin.Context) {
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

	stats, err := (*statsDAO).findByShortID(-1)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, stats)
}