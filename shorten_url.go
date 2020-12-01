// TODO: rename this file accordingly.
package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Showmax/go-fqdn"
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
	_ = c.ShouldBind(&url)

	id, _ := (*dao).save(url)
	shortURL := idToShortURL(id, chars)
	// idGiven := shortURLToID(shortURL, chars)
	// fmt.Printf("Url given: [%d]\n", idGiven)

	fqdn, err := fqdn.FqdnHostname()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	domain := fmt.Sprintf("%s%s", fqdn, serverPort)
	littleuLink := fmt.Sprintf("%s/u/%s", domain, shortURL)

	c.HTML(
		http.StatusOK,
		"url_shorten_summary.html",
		// Pass the data that the page uses
		gin.H{
			"title":        "Home",
			"url":          url.URL,
			"short_url":    shortURL,
			"domain":       domain,
			"littleu_link": littleuLink,
		},
	)
}

func redirectShortURL(c *gin.Context) {
	shortURLParam := c.Param("url")
	if shortURLParam == "" {
		fmt.Println("valiendo verga")
	}

	// fmt.Println("Holis ... ")
	fmt.Printf("This -> [%s]\n", shortURLParam)
	id := shortURLToID(shortURLParam, chars)

	// fmt.Println(id)
	urlFromDB, err := (*dao).findByID(id)
	if err != nil {

	} else {
		c.Redirect(http.StatusMovedPermanently, urlFromDB.URL)
	}
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
