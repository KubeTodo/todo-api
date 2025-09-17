package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// In memory initialized todo for now.
var todos = []Todo{
	{ID: 1, Title: "Learn Go", Done: false},
	{ID: 2, Title: "Set up CI/CD", Done: false},
}

func main() {
	r := gin.Default()

	// GET all todos
	r.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, todos)
	})

	// POST a new todo
	r.POST("/todos", func(c *gin.Context) {
		var newTodo Todo
		if err := c.ShouldBindJSON(&newTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Assign an ID (in a real app, use a database auto-increment)
		newTodo.ID = len(todos) + 1
		todos = append(todos, newTodo)
		c.JSON(http.StatusCreated, newTodo)
	})

	r.Run(":8080")
}
