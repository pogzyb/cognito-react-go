package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"log"
	"net/http"
	"os"
)

var SALT string
var CognitoJWKS jwk.Set

func init() {
	var err error
	SALT = os.Getenv("SECRET_KEY")
	CognitoJWKS, err = loadCognitoJWKs()
	if err != nil {
		log.Fatalf("could not load jwks: %v", err)
	}
}

func Run(addr string) {
	router := gin.New()
	stuff := router.Group("/api-stuff")
	stuff.GET("/wide-open", Unsecure)
	stuff.GET("/top-secret", JWTMiddleware, Secure)
	stuff.GET("/user-info", JWTMiddleware, UserInfo)
	stuff.GET("/authorize", Authorize)
	log.Fatal(router.Run(addr))
}

// Simple wide-open endpoint
func Unsecure(c *gin.Context) {
	payload := gin.H{"message": "OAuth2.0 Apps are cool!"}
	c.JSON(http.StatusOK, payload)
}

// Protected endpoint that returns a secret message for the current user
func Secure(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		log.Println("could not get username from context")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "oops"})
	} else {
		log.Printf("%s is seeing some secret stuff!\n", username)
		// do some database stuff with user
	}
	payload := gin.H{
		"message": fmt.Sprintf("Hey, %s! Teenage mutant ninja turtles are cooler than OAuth.", username)}
	c.JSON(http.StatusOK, payload)
}

// Protected endpoint that returns information about the current user
func UserInfo(c *gin.Context) {
	bearer := c.Request.Header.Get("Authorization")
	if bearer == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "no token"})
		return
	}
	info, err := getUserInfo(bearer)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": *info})
}

// Uses the given authorization code to request tokens from Cognito.
// This endpoint is invoked by the frontend after the user logs in.
func Authorize(c *gin.Context) {
	// get the authorization code from query params
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	// get tokens from Cognito
	tokens, err := requestTokensFromCognito(code)
	if err != nil {
		log.Printf("problem requesting tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "problem requesting tokens"})
		return
	}
	if _, err := verifyTokenString(tokens.Access); err != nil {
		log.Printf("not could verify access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "bad request"})
		return
	}

	// todo:
	// query "user info" from authorization source and
	// take this opportunity to create user if doesn't exist in local DB?

	// set the refresh token as an encrypted httpOnly cookie
	c.SetCookie(
		"refresh_token",
		encrypt(tokens.Refresh, SALT),
		5000,
		"/api-stuff",
		"localhost",
		false,
		true)
	// return the access token to be saved into the browser session by the frontend
	c.JSON(http.StatusAccepted, gin.H{"access_token": tokens.Access})
}

// todo: Refresh expired tokens
func Refresh(c *gin.Context) {
	encrypted := c.Query("refresh")
	if encrypted == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	decrypted := decrypt(encrypted, SALT)
	tokens, err := exchangeRefresh(decrypted)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", err)})
		return
	}
	c.SetCookie(
		"refresh_token",
		encrypt(tokens.Refresh, SALT),
		5000,
		"/api-stuff",
		"localhost",
		false,
		true)
	// return the access token to be saved into the browser session by the frontend
	c.JSON(http.StatusAccepted, gin.H{"access_token": tokens.Access})
}

// todo: Logout
func Logout(c *gin.Context) {}
