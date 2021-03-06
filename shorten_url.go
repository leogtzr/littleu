// c.AbortWithError(http.StatusNotFound, err)
// c.AbortWithStatus(http.StatusNotFound)
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
)

const (
	// Hours24 ...
	Hours24 = time.Hour * 24 * 7
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

func checkUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionID := session.Get("user_logged_in")

		if sessionID == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "unauthorized"})
			c.Abort()
		}
	}
}

func shorturl(c *gin.Context) {
	var url URL
	_ = c.ShouldBind(&url)

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

	id, _ := (*urlDAO).save(url, &userFound)
	shortURL := idToShortURL(id, chars)

	fqdnHostName, err := fqdn.FqdnHostname()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	domain := net.JoinHostPort(fqdnHostName, serverPort)

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

	_, err := (*urlDAO).update(URLID, oldURL, newURL)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
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

	urlFromDB, err := (*urlDAO).findByID(id)
	if err != nil {
		c.HTML(
			http.StatusInternalServerError,
			"error5xx.html",
			gin.H{
				"title":             "Error",
				"error_description": fmt.Sprintf(`Error redirecting to: %s`, shortURLParam),
			},
		)
	} else {
		c.Redirect(http.StatusMovedPermanently, urlFromDB.URL)
	}
}

func login(config *viper.Viper) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				"ErrorMessage": err.Error(),
			})

			return
		}

		exist, err := (*userDAO).userExists(ux.Username)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		if !exist {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "User does not exist",
			})

			return
		}

		match, err := (*userDAO).validateUserAndPassword(ux.Username, ux.Password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		if !match {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Bad credentials",
			})

			return
		}

		user, err := (*userDAO).findByUsername(ux.Username)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		token, err := CreateTokenString(&user, config)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		c.SetCookie("token", token, 3600, "", "", false, true)
		c.Set("is_logged_in", true)

		session := sessions.Default(c)
		session.Set("user_logged_in", user)

		if err := session.Save(); err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": "Error creating user, contact the administrator.",
			})

			return
		}

		c.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"title": "Home",
			},
		)
	}
}

func generateToken(c *gin.Context) {
	type User struct {
		ID       uint64 `json:"id,omitempty"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	dummyUser := User{
		ID:       1,
		Username: "username",
		Password: "password",
	}

	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json provided")

		return
	}
	// compare the user from the request, with the one we defined:
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

func createTokenFromUser(userid string, config *viper.Viper) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(TokenExpirationMinutes).Unix()

	u, _ := uuid.NewV4()
	td.AccessUUID = u.String()

	td.RtExpires = time.Now().Add(Hours24).Unix()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(config.GetString("secret")))
	if err != nil {
		return nil, fmt.Errorf("error creating signed string: %v", err)
	}

	return td, nil
}

func logout(config *viper.Viper) gin.HandlerFunc {
	return func(c *gin.Context) {
		au, err := ExtractTokenMetadata(c.Request, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "unauthorized")

			return
		}

		deleted, delErr := DeleteAuth(au.AccessUUID)
		if delErr != nil || deleted == 0 { // if any goes wrong
			c.JSON(http.StatusUnauthorized, "unauthorized")

			return
		}

		c.JSON(http.StatusOK, "Successfully logged out")
	}

}

// Render one of HTML, JSON or CSV based on the 'Accept' header of the request
// If the header doesn't specify this, HTML is rendered, provided that
// the template name is present.
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
	render(c, gin.H{
		"title": "Login",
	}, "login.html")
}

func showRegistrationPage(c *gin.Context) {
	render(c,
		gin.H{
			"title": "Register",
		}, "register.html",
	)
}

func register(config *viper.Viper) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtain the POSTed username and password values
		username := c.PostForm("username")
		password := c.PostForm("password")

		if err := validateNewUserFields(username, password); err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		hashPassword := hashAndSalt([]byte(password))

		exists, err := (*userDAO).userExists(username)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		if exists {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": "User already exists",
			})

			return
		}

		newUser, err := (*userDAO).addUser(username, hashPassword)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v", err)
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": "Error creating user, contact the administrator.",
			})
		}

		token, err := CreateTokenString(&newUser, config)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": err.Error(),
			})

			return
		}

		c.SetCookie("token", token, 3600, "", "", false, true)
		c.Set("is_logged_in", true)

		session := sessions.Default(c)
		session.Set("user_logged_in", newUser)

		if err := session.Save(); err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": "Error creating user, contact the administrator.",
			})

			return
		}

		render(c, gin.H{
			"title": "Successful registration & Login",
		}, "login-successful.html")
	}
}

func checkSession(c *gin.Context) {
	session := sessions.Default(c)
	userFound := session.Get("user_logged_in")

	fmt.Println("debug session ... ")
	fmt.Println(userFound)
	fmt.Println("debug session ... end")

	c.JSON(http.StatusOK, "OK...")
}

func viewUsers(c *gin.Context) {
	users, err := (*userDAO).findAll()
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, users)
}

func viewURLs(c *gin.Context) {
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

	urls, err := (*urlDAO).findAllByUser(&userFound)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, urls)
}
