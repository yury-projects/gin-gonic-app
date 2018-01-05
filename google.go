package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const header_authorization = "Authorization"

var (
	GoogleAuth *oauth2.Config = &oauth2.Config{
		ClientID:     os.Getenv("GIN_GONIC_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GIN_GONIC_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GIN_GONIC_GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
)

type googleResponseEmail struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type googleResponseName struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type GoogleResponse struct {
	Id       string                `json:"id"`
	Emails   []googleResponseEmail `json:"emails"`
	Name     googleResponseName    `json:"name"`
	Optional interface{}
}

type UserClaims struct {
	User User
	jwt_lib.StandardClaims
}

// Let's make these private
func createOAuthJWT(user *User) (string, error) {
	// Create the token
	var token *jwt_lib.Token

	if user == nil {
		token = jwt_lib.NewWithClaims(jwt_lib.GetSigningMethod("HS256"), jwt_lib.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 2).Unix(),
			Issuer:    "Gin.Gonic.App",
		})
	} else {
		token = jwt_lib.NewWithClaims(jwt_lib.GetSigningMethod("HS256"), UserClaims{*user, jwt_lib.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			Issuer:    "Gin.Gonic.App",
		}})
	}

	// Sign and get the complete encoded token as a string
	return token.SignedString(jwt_oauth_secret_bytes)
}

func validateOAuthJWT(token string) (*jwt_lib.Token, error) {
	return jwt_lib.ParseWithClaims(token, &UserClaims{}, func(token *jwt_lib.Token) (interface{}, error) {

		return jwt_oauth_secret_bytes, nil
	})
}

func CreateGoogleRedirect(c *gin.Context) {

	state, err := createOAuthJWT(nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
	}

	url := GoogleAuth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func JWTMiddleware(c *gin.Context) {

	auth_header := c.GetHeader(header_authorization)

	jwt_token := strings.Split(auth_header, " ")[1]

	token, jwt_err := validateOAuthJWT(jwt_token)

	// If JWT is not valid, abort with Forbidden
	if jwt_err != nil {
		log.Println(jwt_err)

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid state, unable to validate JWT token"})
		return
	}

	fmt.Println(token.Claims.(*UserClaims).User)
}

func GoogleAuthenticated(c *gin.Context) {
	// Validate state parameter first
	state := c.Query("state")

	_, jwt_err := validateOAuthJWT(state)

	// If JWT is not valid, abort with Forbidden
	if jwt_err != nil {
		log.Println(jwt_err)

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid state, unable to validate JWT token"})
		return
	}

	// Otherwise, proceed getting to verifying OAuth response and getting user data
	code := c.Query("code")
	tok, err := GoogleAuth.Exchange(context.TODO(), code)

	if err != nil {
		log.Fatal(err)
	}

	client := GoogleAuth.Client(context.TODO(), tok)
	resp, err := client.Get("https://www.googleapis.com/plus/v1/people/me")

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var google_response GoogleResponse

	// Defining Anonymous struct to parse the response json
	var g_r struct {
		Id     string `json:"id"`
		Emails []struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"emails"`
		Name struct {
			FamilyName string `json:"familyName"`
			GivenName  string `json:"givenName"`
		} `json:"name"`
	}

	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &google_response)
	fmt.Print(google_response)
	//log.Println("body", string(body))

	err = json.Unmarshal(body, &g_r)
	fmt.Print(g_r)

	user := User{rand.Uint64(), g_r.Emails[0].Value, g_r.Emails[0].Value, g_r.Name.GivenName, g_r.Name.FamilyName}

	fmt.Println(user)
	fmt.Println(createOAuthJWT(&user))

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		return
	}

	c.String(http.StatusOK, string(prettyJSON.Bytes()))
}
