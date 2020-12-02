// c.AbortWithError(http.StatusNotFound, err)
// c.AbortWithStatus(http.StatusNotFound)
package main

import (
	"fmt"
	"net"
	"net/http"

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

	fqdn, err := fqdn.FqdnHostname()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	domain := net.JoinHostPort(fqdn, serverPort)

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

func changeLink(c *gin.Context) {
	var url URLChange
	_ = c.ShouldBind(&url)
	fmt.Println(url)
	fmt.Println(url.ShortURL)
	fmt.Println(url.NewURL)
	fmt.Println(shortURLToID(url.NewURL, chars))
	fmt.Println((*dao).findAll())
}

func redirectShortURL(c *gin.Context) {
	shortURLParam := c.Param("url")
	id := shortURLToID(shortURLParam, chars)

	urlFromDB, err := (*dao).findByID(id)
	if err != nil {

	} else {
		c.Redirect(http.StatusMovedPermanently, urlFromDB.URL)
	}
}

func viewUrls(c *gin.Context) {
	urls, err := (*dao).findAll()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, urls)
}
