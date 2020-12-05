// c.AbortWithError(http.StatusNotFound, err)
// c.AbortWithStatus(http.StatusNotFound)
package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func showIndexPage(c *gin.Context) {

	// Call the HTML method of the Context to render a template
	c.HTML(
		http.StatusOK,
		"index.html",
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

	type formUser struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}

	var ux formUser
	if err := c.ShouldBind(&ux); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid data provided")
		return
	}

	if err := validateNewUserFields(ux.Username, ux.Password); err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
		return
	}

	exist, err := userExists(ux.Username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
		return
	}
	if !exist {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "User does not exist"})
		return
	}

	match, _ := validateUserAndPassword(ux.Username, ux.Password)

	if !match {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Bad credentials"})
		return
	}

	// If the user is created, set the token in a cookie and log the user in
	token := generateSessionToken()
	c.SetCookie("token", token, 3600, "", "", false, true)
	c.Set("is_logged_in", true)

	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"title": "Home",
		},
	)

}

// Previously named "login"
func generateToken(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json provided")
		return
	}
	//compare the user from the request, with the one we defined:
	if dummyUser.Username != u.Username || dummyUser.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "please provide valid login details")
		return
	}
	ts, err := CreateToken(dummyUser.ID, envConfig)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	saveErr := CreateAuth(dummyUser.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, saveErr.Error())
	}
	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}
	c.JSON(http.StatusOK, tokens)
}

// Todo ...
type Todo struct {
	UserID uint64 `json:"user_id"`
	Title  string `json:"title"`
}

// TODO: remove the following function and the test route.
// CreateSomething ...
func CreateSomething(c *gin.Context) {
	var td *Todo
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		fmt.Println("Here ... ")
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	userID, err := FetchAuth(tokenAuth)
	if err != nil {
		fmt.Println("Here ... 2")
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	td.UserID = userID
	//you can proceed to save the Todo to a database
	//but we will just return it to the caller here:
	c.JSON(http.StatusCreated, td)
}

func logout(c *gin.Context) {
	au, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	deleted, delErr := DeleteAuth(au.AccessUUID)
	if delErr != nil || deleted == 0 { //if any goes wrong
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, "Successfully logged out")
}

// Render one of HTML, JSON or CSV based on the 'Accept' header of the request
// If the header doesn't specify this, HTML is rendered, provided that
// the template name is present
func render(c *gin.Context, data gin.H, templateName string) {
	loggedInInterface, _ := c.Get("is_logged_in")
	data["is_logged_in"] = loggedInInterface.(bool)

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}

func showLoginPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title": "Login",
	}, "login.html")
}

func showRegistrationPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	render(c,
		gin.H{
			"title": "Register",
		}, "register.html",
	)
}

/*
func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

*/

func register(c *gin.Context) {
	// Obtain the POSTed username and password values
	username := c.PostForm("username")
	password := c.PostForm("password")

	if err := validateNewUserFields(username, password); err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": err.Error()})
		return
	}

	hashPassword := hashAndSalt([]byte(password))

	newUser := NewUser{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		User:      username,
		Password:  hashPassword,
	}

	exists, err := userExists(username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": err.Error()})
		return
	}
	if exists {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": "User already exists"})
		return
	}

	err = createUser(&newUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": "Error creating user, contact the administrator."})
	}

	// If the user is created, set the token in a cookie and log the user in
	token := generateSessionToken()
	c.SetCookie("token", token, 3600, "", "", false, true)
	c.Set("is_logged_in", true)

	render(c, gin.H{
		"title": "Successful registration & Login"}, "login-successful.html")
}

// TODO: move this to another file.
// TODO: make this secure.
func generateSessionToken() string {
	// We're using a random 16 character string as the session token
	// This is NOT a secure way of generating session tokens
	// DO NOT USE THIS IN PRODUCTION
	return strconv.FormatInt(rand.Int63(), 16)
}

// TODO: move this to another file.
// Check if the supplied username is available
func isUsernameAvailable(username string) bool {
	// for _, u := range userList {
	// 	if u.Username == username {
	// 		return false
	// 	}
	// }
	return true
}
