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

func resetTodos() {
	todos = []Todo{
		{ID: 1, Title: "Learn Go", Done: false},
		{ID: 2, Title: "Set up CI/CD", Done: false},
	}
}

func TestGetTodos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	t.Run("Success", func(t *testing.T) {
		resetTodos()
		req, _ := http.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []Todo
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, 2, len(response))
		assert.Equal(t, "Learn Go", response[0].Title)
	})
}

func TestPostTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	t.Run("Success", func(t *testing.T) {
		resetTodos()
		payload := `{"title": "New Todo", "done": false}`
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Todo
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "New Todo", response.Title)
		assert.Equal(t, 3, response.ID)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		resetTodos()
		payload := `{"title": "New Todo", "done": }` // Invalid JSON
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPutTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	t.Run("Success", func(t *testing.T) {
		resetTodos()
		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Todo
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "Updated Todo", response.Title)
		assert.Equal(t, true, response.Done)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		resetTodos()
		payload := `{"title": "Updated Todo", "done": }` // Invalid JSON
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		resetTodos()
		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/999", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeleteTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()

	t.Run("Success", func(t *testing.T) {
		resetTodos()
		req, _ := http.NewRequest("DELETE", "/todos/1", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "Todo deleted", response["message"])
	})

	t.Run("Not Found", func(t *testing.T) {
		resetTodos()
		req, _ := http.NewRequest("DELETE", "/todos/999", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
