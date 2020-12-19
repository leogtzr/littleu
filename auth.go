package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/spf13/viper"
)

const (
	// RequiredNumberOfFieldsInToken ...
	RequiredNumberOfFieldsInToken = 2
	// TokenExpirationMinutes ...
	TokenExpirationMinutes = time.Minute * 15
)

// TokenDetails ...
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// AccessDetails ...
type AccessDetails struct {
	AccessUUID string
	UserID     uint64
}

// CreateTokenString ...
func CreateTokenString(user *interface{}, config *viper.Viper) (string, error) {
	atClaims := jwt.MapClaims{}

	if u, ok := (*user).(UserMongo); ok {
		atClaims["user_id"] = u.ID.Hex()
	} else {
		atClaims["user_id"] = u.ID
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	token, err := at.SignedString([]byte(config.GetString("ACCESS_SECRET")))
	if err != nil {
		return "", fmt.Errorf("error creating signed string: %w", err)
	}

	return token, nil
}

// CreateToken ...
func CreateToken(userid uint64, config *viper.Viper) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(TokenExpirationMinutes).Unix()

	u, _ := uuid.NewV4()
	td.AccessUUID = u.String()

	td.RtExpires = time.Now().Add(Hours24).Unix()
	u, _ = uuid.NewV4()
	td.RefreshUUID = u.String()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(config.GetString("ACCESS_SECRET")))
	if err != nil {
		return nil, fmt.Errorf("error creating signed string: %w", err)
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	td.RefreshToken, err = rt.SignedString([]byte(config.GetString("REFRESH_SECRET")))
	if err != nil {
		return nil, fmt.Errorf("error creating signed string: %w	", err)
	}

	return td, nil
}

// CreateAuth ...
func CreateAuth(userid uint64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := redisClient.Set(td.AccessUUID, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if errAccess != nil {
		return fmt.Errorf("error setting value: %w", errAccess)
	}

	return redisClient.Set(td.RefreshUUID, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
}

// ExtractToken ...
func ExtractToken(r *http.Request) string {
	bearTokenHeader := r.Header.Get("Authorization")
	// normally Authorization the_token_xxx
	bearTokenFields := strings.Split(bearTokenHeader, " ")
	if len(bearTokenFields) == RequiredNumberOfFieldsInToken {
		return bearTokenFields[1]
	}

	return ""
}

// VerifyToken ...
func VerifyToken(r *http.Request, config *viper.Viper) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.GetString("ACCESS_SECRET")), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	return token, nil
}

// TokenValid ...
func TokenValid(r *http.Request, config *viper.Viper) error {
	token, err := VerifyToken(r, config)
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}

	return nil
}

// ExtractTokenMetadata ...
func ExtractTokenMetadata(r *http.Request, config *viper.Viper) (*AccessDetails, error) {
	token, err := VerifyToken(r, config)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}

		userID, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error extrating token meta data: %w", err)
		}

		return &AccessDetails{
			AccessUUID: accessUUID,
			UserID:     userID,
		}, nil
	}

	return nil, err
}

// FetchAuth ...
func FetchAuth(authD *AccessDetails) (uint64, error) {
	userid, err := redisClient.Get(authD.AccessUUID).Result()
	if err != nil {
		return 0, fmt.Errorf("error getting UUID: %w", err)
	}

	userID, _ := strconv.ParseUint(userid, 10, 64)

	return userID, nil
}

// DeleteAuth ...
func DeleteAuth(uuid string) (int64, error) {
	deleted, err := redisClient.Del(uuid).Result()
	if err != nil {
		return 0, fmt.Errorf("error deleting UUID: %w", err)
	}

	return deleted, nil
}

// TokenAuthMiddleware ...
func TokenAuthMiddleware(config *viper.Viper) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := TokenValid(c.Request, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()

			return
		}

		c.Next()
	}
}
