package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *Database {
	// Use an in-memory database for tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the schema
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT false
		)
	`); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Clean up after the test
	t.Cleanup(func() {
		db.Close()
	})

	return &Database{conn: db}
}

func TestGetTodos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := setupTestDB(t)

	// Insert test data using the same database connection
	_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?), (?, ?)",
		"Learn Go", false,
		"Set up CI/CD", false,
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	r := SetupRouter(testDB)

	req, _ := http.NewRequest("GET", "/todos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Todo
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	assert.Equal(t, 2, len(response))
	if len(response) > 0 {
		assert.Equal(t, "Learn Go", response[0].Title)
		assert.Equal(t, "Set up CI/CD", response[1].Title)
	}
}

func TestPostTodo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := setupTestDB(t)
	r := SetupRouter(testDB)

	t.Run("Success", func(t *testing.T) {
		payload := `{"title": "New Todo", "done": false}`
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		assert.Equal(t, "New Todo", response.Title)
		assert.Equal(t, false, response.Done)
		assert.NotZero(t, response.ID)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
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
	testDB := setupTestDB(t)

	// Insert test data using the same database connection
	_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Test Todo", false)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	r := SetupRouter(testDB)

	t.Run("Success", func(t *testing.T) {
		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		assert.Equal(t, "Updated Todo", response.Title)
		assert.Equal(t, true, response.Done)
		assert.Equal(t, 1, response.ID)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		payload := `{"title": "Updated Todo", "done": }` // Invalid JSON
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
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
	testDB := setupTestDB(t)

	// Insert test data using the same database connection
	_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Test Todo", false)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	r := SetupRouter(testDB)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/todos/1", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		assert.Equal(t, "Todo deleted", response["message"])
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/todos/999", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
