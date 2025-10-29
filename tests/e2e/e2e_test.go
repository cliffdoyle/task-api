package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/cliffdoyle/task-api/internal/models" // Adjust import path
)

// waitForServer polls the health check endpoint until the server is ready
func waitForServer(t *testing.T, baseURL string) {
    // Poll for up to 15 seconds for the server to start
    for i := 0; i < 15; i++ {
        resp, err := http.Get(baseURL + "/health")
        if err == nil && resp.StatusCode == http.StatusOK {
            resp.Body.Close()
            fmt.Println("Server is up and running!")
            return
        }
        time.Sleep(time.Second)
    }
    t.Fatal("Server did not start in time")
}

// TestTaskLifecycle performs a full CRUD lifecycle test on the running API
func TestTaskLifecycle(t *testing.T) {
    baseURL := os.Getenv("API_BASE_URL")
    if baseURL == "" {
        baseURL = "http://localhost:8080"
    }

    // Wait for the server to be ready before running tests
    waitForServer(t, baseURL)

    var createdTask models.Task
    client := &http.Client{Timeout: 10 * time.Second}

    // --- 1. Create a Task (POST) ---
    t.Run("Create Task", func(t *testing.T) {
        createReq := models.CreateTaskRequest{
            Title:       "E2E Test Task",
            Description: "A task created during end-to-end testing.",
        }
        body, _ := json.Marshal(createReq)

        resp, err := http.Post(baseURL+"/api/tasks", "application/json", bytes.NewBuffer(body))
        assert.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusCreated, resp.StatusCode)

        err = json.NewDecoder(resp.Body).Decode(&createdTask)
        assert.NoError(t, err)
        assert.NotZero(t, createdTask.ID)
        assert.Equal(t, createReq.Title, createdTask.Title)
        assert.Equal(t, "pending", createdTask.Status)
    })

    // Ensure we have a valid task ID to proceed
    if createdTask.ID == 0 {
        t.Fatal("Task creation failed, cannot proceed with other tests.")
    }

    // --- 2. Get the Task (GET) ---
    t.Run("Get Task", func(t *testing.T) {
        resp, err := http.Get(fmt.Sprintf("%s/api/tasks/%d", baseURL, createdTask.ID))
        assert.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusOK, resp.StatusCode)

        var fetchedTask models.Task
        err = json.NewDecoder(resp.Body).Decode(&fetchedTask)
        assert.NoError(t, err)
        assert.Equal(t, createdTask.ID, fetchedTask.ID)
        assert.Equal(t, createdTask.Title, fetchedTask.Title)
    })

    // --- 3. Update the Task (PUT) ---
    t.Run("Update Task", func(t *testing.T) {
        updateReq := models.UpdateTaskRequest{
            Status: "completed",
        }
        body, _ := json.Marshal(updateReq)

        req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/tasks/%d", baseURL, createdTask.ID), bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")

        resp, err := client.Do(req)
        assert.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusOK, resp.StatusCode)

        var updatedTask models.Task
        err = json.NewDecoder(resp.Body).Decode(&updatedTask)
        assert.NoError(t, err)
        assert.Equal(t, "completed", updatedTask.Status) // Verify status changed
        assert.Equal(t, createdTask.Title, updatedTask.Title) // Verify title did not change
    })

    // --- 4. Delete the Task (DELETE) ---
    t.Run("Delete Task", func(t *testing.T) {
        req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/tasks/%d", baseURL, createdTask.ID), nil)
        resp, err := client.Do(req)
        assert.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    })

    // --- 5. Verify Deletion (GET) ---
    t.Run("Verify Deletion", func(t *testing.T) {
        resp, err := http.Get(fmt.Sprintf("%s/api/tasks/%d", baseURL, createdTask.ID))
        assert.NoError(t, err)
        defer resp.Body.Close()

        // Expect a 404 Not Found since the task was deleted
        assert.Equal(t, http.StatusNotFound, resp.StatusCode)
    })
}
