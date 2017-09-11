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
	"github.com/robfig/cron"
)

// It's ok to leave this secret exposed
const jwt_secret = "JwTsEcReT"
const cron_rss_feed_check_and_notify = "@every 30m"
const header_x_auth_token = "X-Auth-Token"

var (
	jwt_secret_bytes =  []byte(jwt_secret)

	jwt_oauth_secret = os.Getenv("GIN_GONIC_JWT_OAUTH_SECRET")
	jwt_oauth_secret_bytes = []byte(jwt_oauth_secret)

)


type JSONB map[string]interface{}

func Database() *gorm.DB {
	// Open a db connection
	db, err := gorm.Open("postgres", os.Getenv("GIN_GONIC_DATABASE_URL"))

	if err != nil {
		panic("Failed to connect database")
	}

	return db
}

func SignInMiddleware(c *gin.Context) {

	db := Database()

	token := c.GetHeader(header_x_auth_token)

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

func main()  {

	router := gin.Default()

	db := Database()
	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&Token{})
	db.AutoMigrate(&GUID{})

	db.Model(&Token{}).AddIndex("idx_tokens_token", "token")

	v1 := router.Group("/api/v1/todos")
	{
		v1.POST("", CreateTodo)
		v1.GET("", FetchAllTodo)
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
		oauth.GET("/google", CreateGoogleRedirect)
		oauth.GET("/google/authenticated", GoogleAuthenticated)
		oauth.GET("/slack/authenticated", SlackAuthenticated)
	}

	slack := router.Group("/api/v1/slack")
	{
		slack.POST("/command", HandleCommand)
	}

	rss := router.Group("/api/v1/rss")
	{
		rss.GET("", GetLatestRssFeed)
	}

	private_v2 := router.Group("/api/v2/user")

	private_v2.GET("", JWTMiddleware, FetchAllTodo)

	c := cron.New()
	c.AddFunc(cron_rss_feed_check_and_notify, CheckFeedAndNotify)

	c.Start()

	defer c.Stop()

	router.Run()
}