package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Todo represents a to-do item
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetTodos() ([]Todo, error)
	CreateTodo(*Todo) error
	UpdateTodo(int, Todo) (*Todo, error)
	DeleteTodo(int) error
	GetTodoByID(int) (*Todo, error)
	Close() error
}

// Database wraps the SQLite connection and implements DatabaseInterface
type Database struct {
	conn *sql.DB
}

// NewDatabase creates a new Database instance
func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "./todos.db")
	if err != nil {
		return nil, err
	}

	// Create the todos table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT false
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Database{conn: db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// GetTodos returns all todos
func (db *Database) GetTodos() ([]Todo, error) {
	rows, err := db.conn.Query("SELECT id, title, done FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// CreateTodo adds a new todo
func (db *Database) CreateTodo(todo *Todo) error {
	res, err := db.conn.Exec("INSERT INTO todos (title, done) VALUES (?, ?)", todo.Title, todo.Done)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	todo.ID = int(id)
	return nil
}

// UpdateTodo updates an existing todo
func (db *Database) UpdateTodo(id int, updatedTodo Todo) (*Todo, error) {
	_, err := db.conn.Exec("UPDATE todos SET title = ?, done = ? WHERE id = ?", updatedTodo.Title, updatedTodo.Done, id)
	if err != nil {
		return nil, err
	}

	// Return the updated todo with the correct ID
	updatedTodo.ID = id
	return &updatedTodo, nil
}

// DeleteTodo removes a todo
func (db *Database) DeleteTodo(id int) error {
	_, err := db.conn.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

// GetTodoByID finds a todo by ID
func (db *Database) GetTodoByID(id int) (*Todo, error) {
	var todo Todo
	err := db.conn.QueryRow("SELECT id, title, done FROM todos WHERE id = ?", id).Scan(&todo.ID, &todo.Title, &todo.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &todo, nil
}

// SetupRouter initializes and returns the Gin router with all routes
func SetupRouter(db DatabaseInterface) *gin.Engine {
	r := gin.Default()

	r.GET("/todos", func(c *gin.Context) {
		todos, err := db.GetTodos()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, todos)
	})

	r.POST("/todos", func(c *gin.Context) {
		var newTodo Todo
		if err := c.ShouldBindJSON(&newTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.CreateTodo(&newTodo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, newTodo)
	})

	r.PUT("/todos/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var updatedTodo Todo
		if err := c.ShouldBindJSON(&updatedTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existingTodo, err := db.GetTodoByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if existingTodo == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}

		updatedTodo.ID = id
		if _, err := db.UpdateTodo(id, updatedTodo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, updatedTodo)
	})

	r.DELETE("/todos/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		existingTodo, err := db.GetTodoByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if existingTodo == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}

		if err := db.DeleteTodo(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
	})

	return r
}

func main() {
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := SetupRouter(db)
	r.Run(":8080")
}
