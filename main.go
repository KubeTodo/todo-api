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

	r.Run(":8080")
}
