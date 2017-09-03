package main

import (
	"fmt"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"net/http"
	"strconv"
	"database/sql/driver"
	"encoding/json"
	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/jwt"
	"time"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"io/ioutil"
	"bytes"
	"context"
)

var (
	// It's ok to leave this secret exposed
	jwt_secret = "JwTsEcReT"
	jwt_secret_bytes =  []byte(jwt_secret)

	jwt_oauth_secret = os.Getenv("GIN_GONIC_JWT_OAUTH_SECRET")
	jwt_oauth_secret_bytes = []byte(jwt_oauth_secret)

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


type JSONB map[string]interface{}

func Database() *gorm.DB {
	// Open a db connection
	db, err := gorm.Open("postgres", os.Getenv("GIN_GONIC_DATABASE_URL"))

	if err != nil {
		panic("failed to connect database")
	}

	return db
}

func SignInMiddleware(c *gin.Context) {

	db := Database()

	token := c.GetHeader("X-Auth-Token")

	var result Token

	if err := db.Where("token = ?", token).First(&result).Error; err != nil {
		// Error handling...
		c.AbortWithStatus(http.StatusForbidden)
	}

	c.Set("user", result.Data)

	c.Next()
}

type Token struct {
	gorm.Model
	Token string
	Data JSONB `sql:"type:jsonb"`
}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)

	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type Todo struct {
	gorm.Model
	Title		string `json:"title"`
	Completed	int `json:"completed"`
}

type TransformedTodo struct {
	ID        uint `json:"id"`
	Title     string `json:"title"`
	Completed bool `json:"completed"`
}

func CreateTodo(c *gin.Context) {
	completed, _ := strconv.Atoi(c.PostForm("completed"))

	todo := Todo{Title: c.PostForm("title"), Completed: completed};

	db := Database()
	db.Save(&todo)

	c.JSON(http.StatusCreated, gin.H{"status" : http.StatusCreated, "message" : "Todo item created successfully!", "resourceId": todo.ID})
}

func FetchAllTodo(c *gin.Context)  {

	fmt.Println(c.Get("user"))

	var todos []Todo
	var _todos []TransformedTodo

	db := Database()
	db.Limit(2).Find(&todos)

	if len(todos) <= 0 {

		c.JSON(http.StatusNotFound, gin.H{"status" : http.StatusNotFound, "message" : "No todo found!"})

		return
	}

	// Transforms the todos for building a good response
	for _, item := range todos {

		completed := false

		if item.Completed == 1 {
			completed = true
		} else {
			completed = false
		}

		_todos = append(_todos, TransformedTodo{ID: item.ID, Title:item.Title, Completed: completed})
	}

	c.JSON(http.StatusOK, gin.H{"status" : http.StatusOK, "data" : _todos})
}

func CreateJWT(c *gin.Context)  {
	// Create the token
	token := jwt_lib.NewWithClaims(jwt_lib.GetSigningMethod("HS256"), jwt_lib.StandardClaims{
		Id: "Hello.World",
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		Issuer: "Gin.Gonic.App",
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwt_secret_bytes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func GetPrivate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from private"})
}

// Let's make these private
func createOAuthJWT() (string, error)  {
	// Create the token
	token := jwt_lib.NewWithClaims(jwt_lib.GetSigningMethod("HS256"), jwt_lib.StandardClaims{
		Id: "Gin.Gonic.OAuth",
		ExpiresAt: time.Now().Add(time.Second * 120).Unix(),
		Issuer: "Gin.Gonic.OAuth",
	})

	// Sign and get the complete encoded token as a string
	return token.SignedString(jwt_oauth_secret_bytes)
}

func validateOAuthJWT(token string) (*jwt_lib.Token, error) {
	return jwt_lib.Parse(token, func(token *jwt_lib.Token) (interface{}, error) {

		return jwt_oauth_secret_bytes, nil
	})
}

func CreateGoogleRedirect(c *gin.Context) {

	state, err := createOAuthJWT()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
	}

	url := GoogleAuth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
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
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("body", string(body))

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "\t")
	if err != nil {
		log.Println("JSON parse error: ", err)
		return
	}

	c.String(http.StatusOK, string(prettyJSON.Bytes()))
}

func main()  {

	router := gin.Default()

	db := Database()
	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&Token{})

	db.Model(&Token{}).AddIndex("idx_tokens_token", "token")

	v1 := router.Group("/api/v1/todos")
	{
		v1.POST("/", CreateTodo)
		v1.GET("/", FetchAllTodo)
		//v1.GET("/:id", FetchSingleTodo)
		//v1.PUT("/:id", UpdateTodo)
		//v1.DELETE("/:id", DeleteTodo)
	}

	v1.Use(SignInMiddleware)

	public := router.Group("/api/v1/public")

	public.GET("", CreateJWT)

	private := router.Group("/api/v1/private")

	private.Use(jwt.Auth(jwt_secret))

	private.GET("", GetPrivate)

	oauth := router.Group("/api/v1/oauth")
	{
		oauth.GET("/google", SignInMiddleware, CreateGoogleRedirect)
		oauth.GET("/google/authenticated", GoogleAuthenticated)
	}

	router.Run()
}