package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type CognitoUser struct {
	Sub			string		`json:"sub"`
	Email		string		`json:"email"`
	Verified	string		`json:"email_verified"`
	Created		time.Time	`json:"created"`
	Modified	time.Time	`json:"last_modified"`
}

type CognitoTokens struct {
	ID			string	`json:"token_id"`
	Access 		string	`json:"access_token"`
	Refresh 	string	`json:"refresh_token"`
	Expires 	int		`json:"expires_in"`
	Type		string	`json:"token_type"`
	Error		string	`json:"error,omitempty"`
}

// Loads the App's RSA public keys into memory.
// In order to validate a JWTs signature, the private key id issued
// by cognito must be compared to the public key id.
func loadCognitoJWKs() (jwk.Set, error) {
	region := os.Getenv("AWS_COGNITO_REGION")
	pool := os.Getenv("AWS_COGNITO_POOL_ID")
	uri := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, pool)
	keySet, err := jwk.Fetch(context.TODO(), uri)
	if err != nil {
		return nil, err
	}
	return keySet, nil
}

// Retrieves tokens from Cognito using as part of the "Authorization Code Flow"
// DOCS [MAY 2021]:
// https://aws.amazon.com/blogs/mobile/understanding-amazon-cognito-user-pool-oauth-2-0-grants/
func requestTokensFromCognito(code string) (CognitoTokens, error) {
	var cogTokens CognitoTokens
	// construct cognito token endpoint
	domain := os.Getenv("AWS_COGNITO_AUTH_DOMAIN")
	region := os.Getenv("AWS_COGNITO_REGION")
	clientID := os.Getenv("AWS_COGNITO_APP_CLIENT_ID")
	clientSecret := os.Getenv("AWS_COGNITO_APP_CLIENT_SECRET")
	uri := fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/token", domain, region)
	// create authorization header
	authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))
	// create url-encoded payload
	payload := url.Values{}
	payload.Set("grant_type", "authorization_code")
	payload.Set("code", code)
	payload.Set("client_id", clientID)
	payload.Set("redirect_uri", os.Getenv("AWS_COGNITO_REDIRECT_URI")) // must match the User Pool callback
	// send the request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload.Encode()))
	if err != nil {
		return cogTokens, fmt.Errorf("could not make token request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authHeader))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return cogTokens, fmt.Errorf("could not send request: %v", err)
	}
	defer resp.Body.Close()
	// parse the response and return
	_ = json.NewDecoder(resp.Body).Decode(&cogTokens)
	if cogTokens.Error != "" {
		return cogTokens, fmt.Errorf("bad response from cognito: %v", cogTokens.Error)
	}
	return cogTokens, nil
}

// Verifies these things about the token:
// 1. not expired
// 2. the audience (aud) matches the app client id (created in the AWS Cognito User Pool)
// 3. the issuer (iss) matches the user pool
// 4. token_use should be access or id
//
// DOCS [MAY 2021]:
// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
func validClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("could not check claims")
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return fmt.Errorf("token expired")
	}
	if !claims.VerifyIssuer(os.Getenv("AWS_COGNITO_POOL_ISS"), true) {
		return fmt.Errorf("invalid issuer")
	}
	return nil
}

// Parses, verifies, and returns the given token
func verifyTokenString(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if signingMethod := token.Method.Alg(); signingMethod != "RS256" {
			return nil, fmt.Errorf("mismatched signing method: %s", signingMethod)
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing private kid")
		}
		key, ok := CognitoJWKS.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("no match for given kid")
		}
		if err := validClaims(token); err != nil {
			return nil, fmt.Errorf("invalid claims: %v", err)
		}
		var raw interface{}
		return raw, key.Raw(&raw)
	})
	return token, err
}

// Middleware that checks incoming requests for valid JWTs
func JWTMiddleware(c *gin.Context) {
	bearer := c.Request.Header.Get("Authorization")
	bearerSplit := strings.Split(bearer, " ")
	if len(bearerSplit) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not authorized"})
		c.Abort()
	}
	token, err := verifyTokenString(bearerSplit[1])
	if err != nil {
		log.Printf("jwt error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not authorized"})
		c.Abort()
	} else {
		// Bonus (could be separate function or middleware): add current user into context
		c.Set("username", token.Claims.(jwt.MapClaims)["username"])
		c.Next()
	}
}

func getUserInfo(bearer string) (*CognitoUser, error) {
	domain := os.Getenv("AWS_COGNITO_AUTH_DOMAIN")
	region := os.Getenv("AWS_COGNITO_REGION")
	uri := fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/userInfo", domain, region)
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		log.Printf("problem creating http client: %v", err)
	}
	req.Header.Set("Authorization", bearer)
	if resp, err := client.Do(req); err == nil && resp.StatusCode == http.StatusOK {
		var info CognitoUser
		err := json.NewDecoder(resp.Body).Decode(&info)
		if err != nil {
			return nil, fmt.Errorf("problem unmarshalling userinfo: %v", err)
		}
		return &info, nil
	} else {
		log.Println("")
		//buf := make([]byte, resp.ContentLength)
		//_, err := io.ReadFull(resp.Body, buf)
		//log.Println(err, string(buf))
		return nil, fmt.Errorf("bad request")
	}
}

// todo: DRY out token request functions
func exchangeRefresh(token string) (CognitoTokens, error) {
	var cogTokens CognitoTokens
	// construct cognito token endpoint
	domain := os.Getenv("AWS_COGNITO_AUTH_DOMAIN")
	region := os.Getenv("AWS_COGNITO_REGION")
	clientID := os.Getenv("AWS_COGNITO_APP_CLIENT_ID")
	clientSecret := os.Getenv("AWS_COGNITO_APP_CLIENT_SECRET")
	uri := fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/token", domain, region)
	// create authorization header
	authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))
	// create url-encoded payload
	payload := url.Values{}
	payload.Set("grant_type", "refresh_token")
	payload.Set("client_id", clientID)
	payload.Set("refresh_token", token)
	// send the request
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload.Encode()))
	if err != nil {
		return cogTokens, fmt.Errorf("could not make token request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authHeader))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return cogTokens, fmt.Errorf("could not send request: %v", err)
	}
	defer resp.Body.Close()
	// parse the response and return
	_ = json.NewDecoder(resp.Body).Decode(&cogTokens)
	if cogTokens.Error != "" {
		return cogTokens, fmt.Errorf("bad response from cognito: %v", cogTokens.Error)
	}
	return cogTokens, nil
}