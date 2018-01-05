package todo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"

	"github.com/yury-projects/gin-gonic-app/database"
)

type Todo struct {
	gorm.Model
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Completed int    `json:"completed"`
}

type TransformedTodo struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Completed bool   `json:"completed"`
}

type jsonTodo struct {
	Title     string `form:"title" json:"title" binding:"required"`
	Completed *bool  `form:"completed" json:"completed" binding:"exists"` // Stupid validation!
}

// CreateTodo - handler function for creating new todo
func CreateTodo(c *gin.Context) {

	jsonTodo := jsonTodo{}

	if err := c.ShouldBindWith(&jsonTodo, binding.JSON); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "msg": "Validation error, invalid JSON input"})
		return
	}

	todo := Todo{Title: jsonTodo.Title, Status: 1}

	// TODO: Poor conversion logic to bool and int
	if *jsonTodo.Completed {
		todo.Completed = 1
	}

	db := database.Database()
	db.Save(&todo)

	c.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": "Todo item created successfully!", "resourceId": todo.ID})
}

// UpdateTodo - handler function for updating todo by id
func UpdateTodo(c *gin.Context) {
	todoID, err := strconv.Atoi(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "msg": "Validation error, invalid Todo Id"})
		return
	}

	jsonUpdateTodo := struct {
		Title string `json:"title" binding:"required"`
	}{}

	if err := c.ShouldBindWith(&jsonUpdateTodo, binding.JSON); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "msg": "Validation error, invalid JSON input"})
		return
	}

	db := database.Database()
	db.Model(&Todo{}).Where("id = ?", todoID).Update("title", jsonUpdateTodo.Title)

	c.JSON(http.StatusOK, gin.H{})

}

// DeleteTodo - handler function for deleting todo by id
func DeleteTodo(c *gin.Context) {
	todoID, err := strconv.Atoi(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "msg": "Validation error, invalid Todo Id"})
		return
	}

	db := database.Database()
	db.Model(&Todo{}).Where("id = ?", todoID).Update("status", 0)

	c.JSON(http.StatusOK, gin.H{})
}

// FetchAllTodo - handler function for getting all todo-s
func FetchAllTodo(c *gin.Context) {

	fmt.Println(c.Get("user"))

	var todos []Todo
	var _todos = make([]TransformedTodo, 0)

	db := database.Database()
	db.Limit(20).Where("status = ?", 1).Find(&todos)

	// Transforms the todos for building a good response
	for _, item := range todos {

		completed := item.Completed == 1

		_todos = append(_todos, TransformedTodo{ID: item.ID, Title: item.Title, Completed: completed})
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _todos})
}

// ToggleCompleteness - handler function for "completed" toggling of todo by id
func ToggleCompleteness(isCompleted bool) func(*gin.Context) {
	return func(c *gin.Context) {
		todoID, err := strconv.Atoi(c.Param("id"))

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"code": 201, "msg": "Validation error, invalid Todo Id"})
			return
		}

		completed := 0

		if isCompleted {
			completed = 1
		}

		db := database.Database()
		db.Model(&Todo{}).Where("id = ?", todoID).Update("completed", completed)

		c.JSON(http.StatusOK, gin.H{})
	}
}
