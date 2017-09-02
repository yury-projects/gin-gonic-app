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

	fmt.Println(token)

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
	db.Find(&todos)

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

func main()  {

	router := gin.Default()

	router.Use(SignInMiddleware)

	db := Database()
	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&Token{})

	v1 := router.Group("/api/v1/todos")
	{
		v1.POST("/", CreateTodo)
		v1.GET("/", FetchAllTodo)
		//v1.GET("/:id", FetchSingleTodo)
		//v1.PUT("/:id", UpdateTodo)
		//v1.DELETE("/:id", DeleteTodo)
	}

	router.Run()
}