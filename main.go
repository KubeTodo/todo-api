package main

import (
	"net/http"
	"strconv"

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

	// PUT update a todo
	r.PUT("/todos/:id", func(c *gin.Context) {
	    id := c.Param("id")
	    var updatedTodo Todo
	    if err := c.ShouldBindJSON(&updatedTodo); err != nil {
	        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	        return
	    }

	    // Find the todo by ID
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
	})

// Helper function to convert ID string to int
func toInt(s string) int {
    id, err := strconv.Atoi(s)
    if err != nil {
        return 0
    }
    return id
}

	r.Run(":8080")
}
