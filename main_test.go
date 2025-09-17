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
	return m.GetTodosFunc()
}

func (m *MockDatabase) CreateTodo(todo *Todo) error {
	return m.CreateTodoFunc(todo)
}

func (m *MockDatabase) UpdateTodo(id int, todo Todo) (*Todo, error) {
	return m.UpdateTodoFunc(id, todo)
}

func (m *MockDatabase) DeleteTodo(id int) error {
	return m.DeleteTodoFunc(id)
}

func (m *MockDatabase) GetTodoByID(id int) (*Todo, error) {
	return m.GetTodoByIDFunc(id)
}

func (m *MockDatabase) Close() error {
	return m.CloseFunc()
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

func TestDatabaseMethods(t *testing.T) {
	t.Run("GetTodos", func(t *testing.T) {
		testDB := setupTestDB(t)

		// Insert test data
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?), (?, ?)",
			"Todo 1", false,
			"Todo 2", true,
		)
		assert.NoError(t, err)

		todos, err := testDB.GetTodos()
		assert.NoError(t, err)
		assert.Len(t, todos, 2)
		assert.Equal(t, "Todo 1", todos[0].Title)
		assert.Equal(t, false, todos[0].Done)
		assert.Equal(t, "Todo 2", todos[1].Title)
		assert.Equal(t, true, todos[1].Done)
	})

	t.Run("CreateTodo", func(t *testing.T) {
		testDB := setupTestDB(t)

		todo := Todo{Title: "New Todo", Done: false}
		err := testDB.CreateTodo(&todo)
		assert.NoError(t, err)
		assert.NotZero(t, todo.ID)

		// Verify it was inserted
		created, err := testDB.GetTodoByID(todo.ID)
		assert.NoError(t, err)
		assert.Equal(t, "New Todo", created.Title)
	})

	t.Run("UpdateTodo", func(t *testing.T) {
		testDB := setupTestDB(t)

		// Insert initial todo
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "Old Title", false)
		assert.NoError(t, err)

		updated, err := testDB.UpdateTodo(1, Todo{Title: "New Title", Done: true})
		assert.NoError(t, err)
		assert.Equal(t, "New Title", updated.Title)
		assert.Equal(t, true, updated.Done)

		// Verify update
		verify, err := testDB.GetTodoByID(1)
		assert.NoError(t, err)
		assert.Equal(t, "New Title", verify.Title)
		assert.Equal(t, true, verify.Done)
	})

	t.Run("DeleteTodo", func(t *testing.T) {
		testDB := setupTestDB(t)

		// Insert todo
		_, err := testDB.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", "To Delete", false)
		assert.NoError(t, err)

		// Verify it exists
		before, err := testDB.GetTodoByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, before)

		// Delete it
		err = testDB.DeleteTodo(1)
		assert.NoError(t, err)

		// Verify it's gone
		after, err := testDB.GetTodoByID(1)
		assert.NoError(t, err)
		assert.Nil(t, after)
	})

	t.Run("GetTodoByID Not Found", func(t *testing.T) {
		testDB := setupTestDB(t)

		todo, err := testDB.GetTodoByID(999)
		assert.NoError(t, err)
		assert.Nil(t, todo)
	})
}
