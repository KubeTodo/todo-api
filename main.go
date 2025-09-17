package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Todo represents a to-do item
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// In-memory storage for todos
var todos = []Todo{
	{ID: 1, Title: "Learn Go", Done: false},
	{ID: 2, Title: "Set up CI/CD", Done: false},
}

// getTodos handles GET /todos
func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

// postTodo handles POST /todos
func postTodo(c *gin.Context) {
	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Assign an ID
	newTodo.ID = len(todos) + 1
	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

// putTodo handles PUT /todos/:id
func putTodo(c *gin.Context) {
	id := c.Param("id")
	var updatedTodo Todo
	if err := c.ShouldBindJSON(&updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find and update the todo
	var found bool
	for i, todo := range todos {
		if todo.ID == toInt(id) {
			todos[i].Title = updatedTodo.Title
			todos[i].Done = updatedTodo.Done
			found = true
			c.JSON(http.StatusOK, todos[i])
			return
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
	}
}

// deleteTodo handles DELETE /todos/:id
func deleteTodo(c *gin.Context) {
	id := c.Param("id")

	// Find and remove the todo
	for i, todo := range todos {
		if todo.ID == toInt(id) {
			todos = append(todos[:i], todos[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

// SetupRouter initializes and returns the Gin router with all routes
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/todos", getTodos)
	r.POST("/todos", postTodo)
	r.PUT("/todos/:id", putTodo)
	r.DELETE("/todos/:id", deleteTodo)
	return r
}

// Helper function to convert ID string to int
func toInt(s string) int {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return id
}

func main() {
	r := SetupRouter()
	r.Run(":8080")
}
