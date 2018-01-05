package main

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"os"
	"time"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"github.com/yury-projects/gin-gonic-app/database"
	"github.com/yury-projects/gin-gonic-app/slack"
	"github.com/yury-projects/gin-gonic-app/todo"
)

// It's ok to leave this secret exposed
const (
	jwt_secret                     = "JwTsEcReT"
	cron_rss_feed_check_and_notify = "@every 30m"
	header_x_auth_token            = "X-Auth-Token"
)

var (
	jwt_secret_bytes = []byte(jwt_secret)

	jwt_oauth_secret       = os.Getenv("GIN_GONIC_JWT_OAUTH_SECRET")
	jwt_oauth_secret_bytes = []byte(jwt_oauth_secret)
)

type JSONB map[string]interface{}

func SignInMiddleware(c *gin.Context) {

	db := database.Database()

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
	Data  JSONB `sql:"type:jsonb"`
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

func CreateJWT(c *gin.Context) {

	// Create the token
	token := jwt_lib.NewWithClaims(jwt_lib.GetSigningMethod("HS256"), jwt_lib.StandardClaims{
		Id:        "Hello.World",
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		Issuer:    "Gin.Gonic.App",
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwt_secret_bytes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func getPrivate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from private"})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func main() {

	router := gin.Default()

	db := database.Database()
	db.AutoMigrate(&todo.Todo{})
	db.AutoMigrate(&Token{})
	db.AutoMigrate(&slack.GUID{})

	db.Model(&Token{}).AddIndex("idx_tokens_token", "token")

	router.Use(CORSMiddleware())

	todoV1 := router.Group("/api/v1/todos").Use(jwt.Auth(jwt_secret))
	{
		todoV1.POST("", todo.CreateTodo)
		todoV1.GET("", todo.FetchAllTodo)
		//todoV1v1.GET("/:id", FetchSingleTodo)
		todoV1.PUT("/:id", todo.UpdateTodo)
		todoV1.DELETE("/:id", todo.DeleteTodo)
		todoV1.POST("/:id/complete", todo.ToggleCompleteness(true))
		todoV1.POST("/:id/active", todo.ToggleCompleteness(false))
	}

	public := router.Group("/api/v1/public")

	public.GET("", CreateJWT)

	private := router.Group("/api/v1/private")

	private.Use(jwt.Auth(jwt_secret))

	private.GET("", getPrivate)

	oauth := router.Group("/api/v1/oauth")
	{
		oauth.GET("/google", CreateGoogleRedirect)
		oauth.GET("/google/authenticated", GoogleAuthenticated)
		oauth.GET("/slack/authenticated", slack.SlackAuthenticated)
	}

	slackGroup := router.Group("/api/v1/slack")
	{
		slackGroup.POST("/command", slack.AuthenticateCommand, slack.HandleCommand)
	}

	rssGroup := router.Group("/api/v1/rss")
	{
		rssGroup.GET("", slack.GetLatestRssFeed)
	}

	privateV2 := router.Group("/api/v2/user")
	{
		privateV2.GET("", JWTMiddleware, todo.FetchAllTodo)
	}

	c := cron.New()
	c.AddFunc(cron_rss_feed_check_and_notify, slack.CheckFeedAndNotify)

	c.Start()

	defer c.Stop()

	router.Run()
}
