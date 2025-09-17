package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
