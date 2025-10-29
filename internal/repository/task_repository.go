package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/cliffdoyle/task-api/internal/models"
)

// TaskRepository defines the interface for task data operations
type TaskRepository interface {
	Create(task *models.Task) error
	GetByID(id int) (*models.Task, error)
	GetAll() ([]*models.Task, error)
	Update(task *models.Task) error
	Delete(id int) error
}

var ErrTaskNotFound = errors.New("task not found") //export a custom error

// taskRepository is an implementation of TaskRepository that interacts with a SQL database
type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create inserts a new task into the database
func (r *taskRepository) Create(task *models.Task) error {
	query := `
        INSERT INTO tasks (title, description, status, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	return r.db.QueryRow(query, task.Title, task.Description, task.Status).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

// GetByID retrieves a task by its ID from the database
func (r *taskRepository) GetByID(id int) (*models.Task, error) {
	task := &models.Task{}
	query := `SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

// GetAll retrieves all tasks from the database
func (r *taskRepository) GetAll() ([]*models.Task, error) {
	rows, err := r.db.Query(`SELECT id, title, description, status, created_at, updated_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Update modifies an existing task in the database
func (r *taskRepository) Update(task *models.Task) error {
	query := `
        UPDATE tasks
        SET title = $1, description = $2, status = $3, updated_at = NOW()
        WHERE id = $4
        RETURNING updated_at
    `
	return r.db.QueryRow(query, task.Title, task.Description, task.Status, task.ID).
		Scan(&task.UpdatedAt)
}

// Delete removes a task by its ID from the database
func (r *taskRepository) Delete(id int) error {
	result, err := r.db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task with ID %d not found for deletion", id)
	}
	return nil
}
