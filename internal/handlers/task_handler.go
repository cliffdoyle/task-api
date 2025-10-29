package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv" // For converting string ID from URL to int

	"github.com/cliffdoyle/task-api/internal/models"
	"github.com/cliffdoyle/task-api/internal/service"
	"github.com/gorilla/mux" // For routing and path variables
)

// TaskHandler provides HTTP handlers for task-related operations
type TaskHandler struct {
	service service.TaskService
}

// NewTaskHandler creates a new instance of TaskHandler
func NewTaskHandler(service service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// CreateTask handles POST requests to create a new task
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTask handles GET requests to retrieve a single task by ID
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task ID format", http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTask(id)
	if err != nil {
		// Distinguish between "not found" and other errors
		if err.Error() == fmt.Sprintf("task with ID %d not found", id) { // This check relies on the error message from repository.GetByID
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to retrieve task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// GetAllTasks handles GET requests to retrieve all tasks
func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.GetAllTasks()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrieve tasks: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// UpdateTask handles PUT requests to update an existing task by ID
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task ID format", http.StatusBadRequest)
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(id, &req)
	if err != nil {
		// Distinguish between "not found", "invalid status", and other errors
		if err.Error() == fmt.Sprintf("task with ID %d not found: sql: no rows in result set", id) || err.Error() == fmt.Sprintf("task with ID %d not found", id) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if err.Error() == "invalid status value" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE requests to remove a task by ID
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task ID format", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTask(id)
	if err != nil {
		if err.Error() == fmt.Sprintf("task with ID %d not found for deletion", id) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}
