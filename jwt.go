package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	getenvs "gitlab.com/avarf/getenvs"
)

// Secret key to uniquely sign the token
var key []byte

// Credential User's login information
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Token jwt Standard Claim Object
type Token struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Create a dummy local db instance as a key value pair
var userdb = map[string]string{
	"jwtuser1": "pa$$W0rd1",
}

// assign the secret key to key variable on program's first run
func init() {
	// read the secret_key from the environment variables
	key = []byte(getenvs.GetEnvString("MAGOVER_JWT_SECRET_KEY", "james8ond"))
}

// ---- swagger Informations
// @Tags         JWT
// @router /v1/jwt/login [post]
// @summary get JWT token
// @consume application/json
// @param data body Credential false "Your credential"
// @produce application/json
// @success 200 string OK
// @failure 401 string Unauthorized
// @failure 500 string Internal Server Error
func jwtLoginHandler(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	// create a Credentials object
	var creds Credential
	// decode json to struct
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// verify if user exist or not
	userPassword, ok := userdb[creds.Username]

	// if user exist, verify the password
	if !ok || userPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a token object and add the Username and StandardClaims
	var tokenClaim = Token{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// Enter expiration in milisecond
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
	}

	// Create a new claim with HS256 algorithm and token claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaim)

	tokenString, err := token.SignedString(key)

	if err != nil {
		log.Printf("[JWT/LOGIN] ERROR: %s", err)
	}
	json.NewEncoder(w).Encode(tokenString)
}

// ---- swagger Informations
// @Tags         JWT
// @router /v1/jwt/test [get]
// @summary Test JWT token
// @security BearerAuth
// @success 200 string OK
// @failure 401 string Unauthorized
// @failure 500 string Internal Server Error
func jwtTestHandler(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request

	// get the bearer token from the reuest header
	bearerToken := r.Header.Get("Authorization")

	// validate token, it will return Token and error
	token, err := ValidateToken(bearerToken)

	if err != nil {
		// check if Error is Signature Invalid Error
		if err == jwt.ErrSignatureInvalid {
			// return the Unauthorized Status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Return the Bad Request for any other error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Validate the token if it expired or not
	if !token.Valid {
		// return the Unauthoried Status for expired token
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Type cast the Claims to *Token type
	user := token.Claims.(*Token)

	// send the username Dashboard message
	//json.NewEncoder(w).Encode(fmt.Sprintf("%s Dashboard", user.Username))
	c.JSON(http.StatusOK, gin.H{
		"message":    "Welcome " + user.Username,
		"expiration": time.Unix(user.ExpiresAt, 0),
	})
}

// ValidateToken validates the token with the secret key and return the object
func ValidateToken(bearerToken string) (*jwt.Token, error) {

	// format the token string
	tokenString := strings.Split(bearerToken, " ")[1]

	// Parse the token with tokenObj
	token, err := jwt.ParseWithClaims(tokenString, &Token{}, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	// return token and err
	return token, err
}
