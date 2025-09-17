package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetTodos(t *testing.T) {
	// Open a server in test mode
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, todos)
	})

	req, err := http.NewRequest("GET", "/todos", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Record the response
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Decode the response
	var response []Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Assert the response
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "Learn Go", response[0].Title)
	assert.Equal(t, "Set up CI/CD", response[1].Title)
}

func TestPostTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/todos", func(c *gin.Context) {
		var newTodo Todo
		if err := c.ShouldBindJSON(&newTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newTodo.ID = len(todos) + 1
		todos = append(todos, newTodo)
		c.JSON(http.StatusCreated, newTodo)
	})

	// Create a JSON payload for the new todo
	payload := `{"title": "Test Todo", "done": false}`

	// Create a request with the payload
	req, err := http.NewRequest("POST", "/todos", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusCreated, w.Code)

	// Decode the response
	var response Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Assert the response
	assert.Equal(t, "Test Todo", response.Title)
	assert.Equal(t, false, response.Done)
	assert.Equal(t, 3, response.ID)
}

func TestPutTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PUT("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		var updatedTodo Todo
		if err := c.ShouldBindJSON(&updatedTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

	// Create a JSON payload for the updated todo
	payload := `{"title": "Updated Todo", "done": true}`

	// Create a request with the payload
	req, err := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Decode the response
	var response Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Assert the response
	assert.Equal(t, "Updated Todo", response.Title)
	assert.Equal(t, true, response.Done)
	assert.Equal(t, 1, response.ID)
}

func TestDeleteTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")

		for i, todo := range todos {
			if todo.ID == toInt(id) {
				todos = append(todos[:i], todos[i+1:]...)
				c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
	})

	// Create a request to delete todo with ID 1
	req, err := http.NewRequest("DELETE", "/todos/1", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Decode the response
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Assert the response
	assert.Equal(t, "Todo deleted", response["message"])
}
