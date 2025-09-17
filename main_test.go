package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// MockDatabase implements DatabaseInterface for testing
type MockDatabase struct {
	GetTodosFunc    func() ([]Todo, error)
	CreateTodoFunc  func(*Todo) error
	UpdateTodoFunc  func(int, Todo) (*Todo, error)
	DeleteTodoFunc  func(int) error
	GetTodoByIDFunc func(int) (*Todo, error)
	CloseFunc       func() error
}

func (m *MockDatabase) GetTodos() ([]Todo, error) {
	if m.GetTodosFunc != nil {
		return m.GetTodosFunc()
	}
	return nil, errors.New("GetTodosFunc not implemented")
}

func (m *MockDatabase) CreateTodo(todo *Todo) error {
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(todo)
	}
	return errors.New("CreateTodoFunc not implemented")
}

func (m *MockDatabase) UpdateTodo(id int, todo Todo) (*Todo, error) {
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(id, todo)
	}
	return nil, errors.New("UpdateTodoFunc not implemented")
}

func (m *MockDatabase) DeleteTodo(id int) error {
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(id)
	}
	return errors.New("DeleteTodoFunc not implemented")
}

func (m *MockDatabase) GetTodoByID(id int) (*Todo, error) {
	if m.GetTodoByIDFunc != nil {
		return m.GetTodoByIDFunc(id)
	}
	return nil, errors.New("GetTodoByIDFunc not implemented")
}

func (m *MockDatabase) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return errors.New("CloseFunc not implemented")
}

func setupTestDB(t *testing.T) *Database {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT false
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return &Database{conn: db}
}

func TestGetTodos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)

		// Insert test data
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?), (?, ?)",
			"Learn Go", false,
			"Set up CI/CD", false,
		)
		assert.NoError(t, err)

		r := SetupRouter(testDB)

		req, _ := http.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []Todo
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
	})

	t.Run("Database Error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			GetTodosFunc: func() ([]Todo, error) {
				return nil, errors.New("database error")
			},
		}

		r := SetupRouter(mockDB)

		req, _ := http.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPostTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		payload := `{"title": "New Todo", "done": false}`
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response Todo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "New Todo", response.Title)
		assert.Equal(t, false, response.Done)
		assert.NotZero(t, response.ID)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		payload := `{"title": "New Todo", "done": }`
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database Error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			CreateTodoFunc: func(*Todo) error {
				return errors.New("database error")
			},
		}
		r := SetupRouter(mockDB)

		payload := `{"title": "New Todo", "done": false}`
		req, _ := http.NewRequest("POST", "/todos", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPutTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)

		// Insert test data
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Test Todo", false)
		assert.NoError(t, err)

		r := SetupRouter(testDB)

		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Todo
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Todo", response.Title)
		assert.Equal(t, true, response.Done)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/abc", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		payload := `{"title": "Updated Todo", "done": }`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/999", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Database Error on Get", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			GetTodoByIDFunc: func(int) (*Todo, error) {
				return nil, errors.New("database error")
			},
			UpdateTodoFunc: func(int, Todo) (*Todo, error) {
				return nil, errors.New("should not be called")
			},
		}
		r := SetupRouter(mockDB)

		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Database Error on Update", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			GetTodoByIDFunc: func(int) (*Todo, error) {
				return &Todo{ID: 1, Title: "Old", Done: false}, nil
			},
			UpdateTodoFunc: func(int, Todo) (*Todo, error) {
				return nil, errors.New("database error")
			},
		}
		r := SetupRouter(mockDB)

		payload := `{"title": "Updated Todo", "done": true}`
		req, _ := http.NewRequest("PUT", "/todos/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)

		// Insert test data
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Test Todo", false)
		assert.NoError(t, err)

		r := SetupRouter(testDB)

		req, _ := http.NewRequest("DELETE", "/todos/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Todo deleted", response["message"])
	})

	t.Run("Invalid ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		req, _ := http.NewRequest("DELETE", "/todos/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		testDB := setupTestDB(t)
		r := SetupRouter(testDB)

		req, _ := http.NewRequest("DELETE", "/todos/999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Database Error on Get", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			GetTodoByIDFunc: func(int) (*Todo, error) {
				return nil, errors.New("database error")
			},
			DeleteTodoFunc: func(int) error {
				return errors.New("should not be called")
			},
		}
		r := SetupRouter(mockDB)

		req, _ := http.NewRequest("DELETE", "/todos/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Database Error on Delete", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockDB := &MockDatabase{
			GetTodoByIDFunc: func(int) (*Todo, error) {
				return &Todo{ID: 1, Title: "Test", Done: false}, nil
			},
			DeleteTodoFunc: func(int) error {
				return errors.New("database error")
			},
		}
		r := SetupRouter(mockDB)

		req, _ := http.NewRequest("DELETE", "/todos/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDatabase_CreateTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := setupTestDB(t)
		todo := Todo{Title: "Test Create", Done: false}

		err := db.CreateTodo(&todo)
		assert.NoError(t, err)
		assert.NotZero(t, todo.ID)

		// Verify the todo was inserted
		created, err := db.GetTodoByID(todo.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Test Create", created.Title)
		assert.Equal(t, false, created.Done)
	})

	t.Run("Error", func(t *testing.T) {
		db := setupTestDB(t)
		db.Close() // Force an error

		todo := Todo{Title: "Test Create", Done: false}
		err := db.CreateTodo(&todo)
		assert.Error(t, err)
	})
}

func TestDatabase_GetTodos(t *testing.T) {
	t.Run("With Data", func(t *testing.T) {
		db := setupTestDB(t)

		// Insert test data
		_, err := db.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?), (?, ?)",
			"Todo 1", false,
			"Todo 2", true,
		)
		assert.NoError(t, err)

		todos, err := db.GetTodos()
		assert.NoError(t, err)
		assert.Len(t, todos, 2)
		assert.Equal(t, "Todo 1", todos[0].Title)
		assert.Equal(t, false, todos[0].Done)
		assert.Equal(t, "Todo 2", todos[1].Title)
		assert.Equal(t, true, todos[1].Done)
	})

	t.Run("Empty", func(t *testing.T) {
		db := setupTestDB(t)
		todos, err := db.GetTodos()
		assert.NoError(t, err)
		assert.Empty(t, todos)
	})

	t.Run("Error", func(t *testing.T) {
		db := setupTestDB(t)
		db.Close() // Force an error

		_, err := db.GetTodos()
		assert.Error(t, err)
	})
}

func TestDatabase_UpdateTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := setupTestDB(t)

		// Insert initial todo
		_, err := db.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Old Title", false)
		assert.NoError(t, err)

		updatedTodo := Todo{Title: "New Title", Done: true}
		result, err := db.UpdateTodo(1, updatedTodo)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Title", result.Title)
		assert.Equal(t, true, result.Done)

		// Verify the update
		todo, err := db.GetTodoByID(1)
		assert.NoError(t, err)
		assert.Equal(t, "New Title", todo.Title)
		assert.Equal(t, true, todo.Done)
	})

	t.Run("Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		updatedTodo := Todo{Title: "New Title", Done: true}
		result, err := db.UpdateTodo(999, updatedTodo)
		assert.NoError(t, err)
		assert.Nil(t, result) // Expect nil for not found
	})

	t.Run("Error", func(t *testing.T) {
		db := setupTestDB(t)
		db.Close() // Force an error

		updatedTodo := Todo{Title: "New Title", Done: true}
		_, err := db.UpdateTodo(1, updatedTodo)
		assert.Error(t, err)
	})
}

func TestDatabase_DeleteTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := setupTestDB(t)

		// Insert todo
		_, err := db.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "To Delete", false)
		assert.NoError(t, err)

		// Verify it exists
		before, err := db.GetTodoByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, before)

		// Delete it
		err = db.DeleteTodo(1)
		assert.NoError(t, err)

		// Verify it's gone
		after, err := db.GetTodoByID(1)
		assert.NoError(t, err)
		assert.Nil(t, after)
	})

	t.Run("Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		err := db.DeleteTodo(999)
		assert.NoError(t, err) // SQLite won't return error for deleting non-existent row
	})

	t.Run("Error", func(t *testing.T) {
		db := setupTestDB(t)
		db.Close() // Force an error

		err := db.DeleteTodo(1)
		assert.Error(t, err)
	})
}

func TestDatabase_GetTodoByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := setupTestDB(t)

		// Insert test data
		_, err := db.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Test Todo", false)
		assert.NoError(t, err)

		todo, err := db.GetTodoByID(1)
		assert.NoError(t, err)
		assert.Equal(t, "Test Todo", todo.Title)
		assert.Equal(t, false, todo.Done)
	})

	t.Run("Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		todo, err := db.GetTodoByID(999)
		assert.NoError(t, err)
		assert.Nil(t, todo)
	})

	t.Run("Error", func(t *testing.T) {
		db := setupTestDB(t)
		db.Close() // Force an error

		_, err := db.GetTodoByID(1)
		assert.Error(t, err)
	})
}
