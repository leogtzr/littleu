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

	// Call the HTML method of the Context to render a template
	c.HTML(
		http.StatusOK,
		// Use the index.html template
		"index.html",
		// Pass the data that the page uses
		gin.H{
			"title": "Home",
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

func debugURLSIDs(urls ...string) {
	for _, url := range urls {
		id := shortURLToID(url, chars)
		fmt.Printf("The id for '%s' is %d\n", url, id)
	}
}

func changeLink(c *gin.Context) {
	var url URLChange
	_ = c.ShouldBind(&url)

	debugURLSIDs(url.NewURL, url.ShortURL)

	URLID := shortURLToID(url.ShortURL, chars)

	oldURL := URL{
		URL: url.ShortURL,
	}

	newURL := URL{
		URL: url.NewURL,
	}

	_, err := (*dao).update(URLID, oldURL, newURL)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.HTML(
		http.StatusOK,
		"littleu_linkchanged.html",
		// Pass the data that the page uses
		gin.H{
			"title":     "littleu - link changed",
			"from_link": url.ShortURL,
			"to_link":   url.NewURL,
		},
	)

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

func login(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	//compare the user from the request, with the one we defined:
	if user.Username != u.Username || user.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "Please provide valid login details")
		return
	}
	ts, err := CreateToken(user.ID, envConfig)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	saveErr := CreateAuth(user.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, saveErr.Error())
	}
	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}
	c.JSON(http.StatusOK, tokens)
}
