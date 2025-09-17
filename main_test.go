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
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	req, err := http.NewRequest("GET", "/todos", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	assert.Equal(t, 2, len(response))
	assert.Equal(t, "Learn Go", response[0].Title)
	assert.Equal(t, "Set up CI/CD", response[1].Title)
}

func TestPostTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	payload := `{"title": "Test Todo", "done": false}`
	req, err := http.NewRequest("POST", "/todos", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	assert.Equal(t, "Test Todo", response.Title)
	assert.Equal(t, false, response.Done)
	assert.Equal(t, 3, response.ID)
}

func TestPutTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	payload := `{"title": "Updated Todo", "done": true}`
	req, err := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	assert.Equal(t, "Updated Todo", response.Title)
	assert.Equal(t, true, response.Done)
	assert.Equal(t, 1, response.ID)
}

func TestDeleteTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	req, err := http.NewRequest("DELETE", "/todos/1", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	assert.Equal(t, "Todo deleted", response["message"])
}
