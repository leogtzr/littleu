// TODO: rename this file accordingly.
package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func showIndexPage(c *gin.Context) {
	articles := getAllArticles()

	// Call the HTML method of the Context to render a template
	c.HTML(
		http.StatusOK,
		// Use the index.html template
		"index.html",
		// Pass the data that the page uses
		gin.H{
			"title":   "Home",
			"payload": articles,
		},
	)
}

func shorturl(c *gin.Context) {

	var url URL
	if c.ShouldBind(&url) == nil {
		fmt.Println(url)
		fmt.Println(":)")
	}

	fmt.Println("hmmm, I am here ... ")

	c.HTML(
		http.StatusOK,
		"url_shorten_summary.html",
		// Pass the data that the page uses
		gin.H{
			"title": "Home",
			"url":   url.URL,
		},
	)
}

func getArticle(c *gin.Context) {
	// Check if the article ID is valid
	if articleID, err := strconv.Atoi(c.Param("article_id")); err == nil {
		// Check if the article exists
		if article, err := getArticleByID(articleID); err == nil {
			// Call the HTML method of the Context to render a template
			c.HTML(
				// Set the HTTP status to 200 (OK)
				http.StatusOK,
				// Use the index.html template
				"article.html",
				// Pass the data that the page uses
				gin.H{
					"title":   article.Title,
					"payload": article,
				},
			)

		} else {
			// If the article is not found, abort with an error
			c.AbortWithError(http.StatusNotFound, err)
		}

	} else {
		// If an invalid article ID is specified in the URL, abort with an error
		c.AbortWithStatus(http.StatusNotFound)
	}
}
