package main

import (
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
		if shortURLParam == "" {
			c.HTML(
				http.StatusInternalServerError,
				"error5xx.html",
				gin.H{
					"title":             "Error",
					"error_description": `error: missing url argument to redirect to`,
				},
			)

			return
		}

		session := sessions.Default(c)
		user := session.Get("user_logged_in")

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "unauthorized"})
			c.Abort()
		}

		headers := map[string][]string(c.Request.Header)
		// As of now, all the headers are being saved, we might want to consider to save only a few, such as:
		// Referrer, User-Agent, etc
		(*statsDAO).save(shortURLParam, &headers, &user)
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

	stats, err := (*statsDAO).findAllByUser(&userFound)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, stats)
}