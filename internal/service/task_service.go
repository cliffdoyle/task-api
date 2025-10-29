package service

import (
	"errors"
	"fmt"

	"github.com/cliffdoyle/task-api/internal/models"
	"github.com/cliffdoyle/task-api/internal/repository"
)

// TaskService defines the interface for task-related business logic
type TaskService interface {
	CreateTask(req *models.CreateTaskRequest) (*models.Task, error)
	GetTask(id int) (*models.Task, error)
	GetAllTasks() ([]*models.Task, error)
	UpdateTask(id int, req *models.UpdateTaskRequest) (*models.Task, error)
	DeleteTask(id int) error
}

// taskService is an implementation of TaskService
type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService creates a new instance of TaskService
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// CreateTask handles the creation of a new task, including validation
func (s *taskService) CreateTask(req *models.CreateTaskRequest) (*models.Task, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending", // Default status for new tasks
	}

	if err := s.repo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task in repository: %w", err)
	}

	return task, nil
}

// GetTask retrieves a single task by its ID
func (s *taskService) GetTask(id int) (*models.Task, error) {
	if id <= 0 {
		return nil, errors.New("invalid task ID")
	}
	task, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from repository: %w", err)
	}
	return task, nil
}

// GetAllTasks retrieves all tasks
func (s *taskService) GetAllTasks() ([]*models.Task, error) {
	tasks, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks from repository: %w", err)
	}
	return tasks, nil
}

// UpdateTask updates an existing task with the provided request data
func (s *taskService) UpdateTask(id int, req *models.UpdateTaskRequest) (*models.Task, error) {
	if id <= 0 {
		return nil, errors.New("invalid task ID")
	}

	existingTask, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("task with ID %d not found: %w", id, err)
	}

	// Apply updates if fields are provided
	if req.Title != "" {
		existingTask.Title = req.Title
	}
	if req.Description != "" {
		existingTask.Description = req.Description
	}
	if req.Status != "" {
		// Basic validation for status
		if req.Status != "pending" && req.Status != "in_progress" && req.Status != "completed" {
			return nil, errors.New("invalid status value")
		}
		existingTask.Status = req.Status
	}

	if err := s.repo.Update(existingTask); err != nil {
		return nil, fmt.Errorf("failed to update task in repository: %w", err)
	}

	return existingTask, nil
}

// DeleteTask deletes a task by its ID
func (s *taskService) DeleteTask(id int) error {
	if id <= 0 {
		return errors.New("invalid task ID")
	}
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete task from repository: %w", err)
	}
	return nil
}
