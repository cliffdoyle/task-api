package models

import "time"

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"` // "pending", "in_progress", "completed"
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
}

type UpdateTaskRequest struct {
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Status      string `json:"status,omitempty"`
}