package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cliffdoyle/task-api/internal/handlers"
	"github.com/cliffdoyle/task-api/internal/models"
	"github.com/cliffdoyle/task-api/internal/repository"
	"github.com/cliffdoyle/task-api/internal/service"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
)

// setupTestDB connects to the test database and ensures it's clean
func setupTestDB(t *testing.T) *sql.DB {
	// We expect the local Dockerized PostgreSQL to be running.
	// Use an environment variable for connection, similar to main.go.
	// For integration tests, we'll often use a dedicated test database
	// or ensure transactions are rolled back. For simplicity here,
	// we'll truncate tables.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback for direct test run, but better to set via script
		log.Println("DATABASE_URL not set for integration tests, defaulting to local postgres:5432")
		dbURL = "postgresql://taskapi:taskapi123@localhost:5432/taskapi?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Ping to ensure connection is live
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Clean up tables before each test suite or potentially before each test
	// For simplicity, we'll truncate all tables. In a real-world scenario,
	// you might use test transactions or dedicated test databases for isolation.
	_, err = db.Exec(`TRUNCATE TABLE tasks RESTART IDENTITY CASCADE;`)
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	log.Println("Test database setup complete and cleaned.")
	return db
}

// setupRouter initializes the application's router and handlers for testing
func setupRouter(db *sql.DB) *mux.Router {
	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	r := mux.NewRouter()
	r.HandleFunc("/api/tasks", taskHandler.CreateTask).Methods("POST")
	r.HandleFunc("/api/tasks", taskHandler.GetAllTasks).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.GetTask).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	r.HandleFunc("/api/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")
	r.HandleFunc("/health", healthCheck).Methods("GET") // Health check for integration sanity
	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Helper to make requests and decode responses
func executeRequest(router *mux.Router, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// TestCreateTaskIntegration verifies task creation via API
func TestCreateTaskIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close() // Close DB connection after test suite
	router := setupRouter(db)

	reqBody := models.CreateTaskRequest{
		Title:       "Integration Test Task",
		Description: "Testing integration for task creation",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected HTTP 201 Created")

	var task models.Task
	err := json.NewDecoder(rr.Body).Decode(&task)
	assert.NoError(t, err)
	assert.Equal(t, "Integration Test Task", task.Title)
	assert.NotZero(t, task.ID) // Check that an ID was assigned by the DB
}

// TestGetAllTasksIntegration verifies fetching all tasks
func TestGetAllTasksIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router := setupRouter(db)

	// First, create a few tasks directly in the DB for known state
	_, err := db.Exec(`INSERT INTO tasks (title, description, status, created_at, updated_at) VALUES 
        ('Task One', 'Desc One', 'pending', NOW(), NOW());`)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Add a tiny delay to ensure timestamps differ

	_, err = db.Exec(`INSERT INTO tasks (title, description, status, created_at, updated_at) VALUES 
        ('Task Two', 'Desc Two', 'completed', NOW(), NOW());`)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/tasks", nil)
	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK")

	var tasks []*models.Task
	err = json.NewDecoder(rr.Body).Decode(&tasks)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2, "Expected 2 tasks")
	assert.Equal(t, "Task Two", tasks[0].Title) // Assuming default sort by created_at DESC from repo
	assert.Equal(t, "Task One", tasks[1].Title)
}

// TestGetTaskByIDIntegration verifies fetching a single task
func TestGetTaskByIDIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router := setupRouter(db)

	// Insert a specific task for retrieval
	var taskID int
	err := db.QueryRow(`INSERT INTO tasks (title, description, status, created_at, updated_at) VALUES 
        ('Specific Task', 'Specific Description', 'pending', NOW(), NOW()) RETURNING id;`).Scan(&taskID)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/tasks/%d", taskID), nil)
	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK")

	var task models.Task
	err = json.NewDecoder(rr.Body).Decode(&task)
	assert.NoError(t, err)
	assert.Equal(t, "Specific Task", task.Title)
	assert.Equal(t, taskID, task.ID)
}

// TestGetTaskByID_NotFound verifies fetching a non-existent task
func TestGetTaskByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router := setupRouter(db)

	req := httptest.NewRequest("GET", "/api/tasks/999", nil) // Non-existent ID
	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected HTTP 404 Not Found")
	bodyBytes, _ := io.ReadAll(rr.Body)
	assert.Equal(t, "task not found\n", string(bodyBytes), "Expected simple 'task not found' message")
}

// TestUpdateTaskIntegration verifies updating a task
func TestUpdateTaskIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router := setupRouter(db)

	var taskID int
	err := db.QueryRow(`INSERT INTO tasks (title, description, status, created_at, updated_at) VALUES 
        ('Task to Update', 'Old Description', 'pending', NOW(), NOW()) RETURNING id;`).Scan(&taskID)
	assert.NoError(t, err)

	updateReqBody := models.UpdateTaskRequest{
		Description: "New Description",
		Status:      "completed",
	}
	body, _ := json.Marshal(updateReqBody)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/tasks/%d", taskID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK")

	var updatedTask models.Task
	err = json.NewDecoder(rr.Body).Decode(&updatedTask)
	assert.NoError(t, err)
	assert.Equal(t, taskID, updatedTask.ID)
	assert.Equal(t, "Task to Update", updatedTask.Title) // Title unchanged as not sent
	assert.Equal(t, "New Description", updatedTask.Description)
	assert.Equal(t, "completed", updatedTask.Status)

	// Verify in DB directly
	var dbStatus string
	err = db.QueryRow("SELECT status FROM tasks WHERE id = $1", taskID).Scan(&dbStatus)
	assert.NoError(t, err)
	assert.Equal(t, "completed", dbStatus)
}

// TestDeleteTaskIntegration verifies deleting a task
func TestDeleteTaskIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router := setupRouter(db)

	var taskID int
	err := db.QueryRow(`INSERT INTO tasks (title, description, status, created_at, updated_at) VALUES 
        ('Task to Delete', 'Delete me', 'pending', NOW(), NOW()) RETURNING id;`).Scan(&taskID)
	assert.NoError(t, err)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/tasks/%d", taskID), nil)
	rr := executeRequest(router, req)

	assert.Equal(t, http.StatusNoContent, rr.Code, "Expected HTTP 204 No Content")

	// Verify task is deleted from DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Expected task to be deleted from DB")
}
