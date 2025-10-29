package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/cliffdoyle/task-api/internal/models"
)

// MockTaskRepository is a mock implementation of the TaskRepository interface.
// This allows us to test the service layer without actually touching a database.
type MockTaskRepository struct {
	mock.Mock
}

// Create mocks the Create method of the repository
func (m *MockTaskRepository) Create(task *models.Task) error {
	args := m.Called(task)
	// Simulate setting ID and timestamps as a real DB would
	if args.Error(0) == nil {
		task.ID = 1 // Assign a dummy ID
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

// GetByID mocks the GetByID method of the repository
func (m *MockTaskRepository) GetByID(id int) (*models.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil { // If no task is returned (e.g., not found)
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

// GetAll mocks the GetAll method of the repository
func (m *MockTaskRepository) GetAll() ([]*models.Task, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Task), args.Error(1)
}

// Update mocks the Update method of the repository
func (m *MockTaskRepository) Update(task *models.Task) error {
	args := m.Called(task)
	if args.Error(0) == nil {
		task.UpdatedAt = time.Now() // Simulate DB update
	}
	return args.Error(0)
}

// Delete mocks the Delete method of the repository
func (m *MockTaskRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- Test Cases for CreateTask ---
func TestCreateTask_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	req := &models.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	// Expect the Create method to be called once with any Task object
	// and return no error.
	mockRepo.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)

	// Act
	task, err := service.CreateTask(req)

	// Assert
	assert.NoError(t, err) // Expect no error
	assert.NotNil(t, task) // Expect a task object to be returned
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "pending", task.Status) // Verify default status
	mockRepo.AssertExpectations(t)          // Verify that the expected mock calls happened
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	req := &models.CreateTaskRequest{
		Title:       "", // Empty title
		Description: "Test Description",
	}

	// The repository's Create method should NOT be called because validation fails first.
	mockRepo.AssertNotCalled(t, "Create")

	// Act
	task, err := service.CreateTask(req)

	// Assert
	assert.Error(t, err) // Expect an error
	assert.Nil(t, task)  // Expect no task to be returned
	assert.Equal(t, "title is required", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestCreateTask_RepoError(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	req := &models.CreateTaskRequest{
		Title:       "Failing Task",
		Description: "This task will fail at the repo layer.",
	}

	// Configure mock to return an error when Create is called
	repoError := errors.New("database connection failed")
	mockRepo.On("Create", mock.AnythingOfType("*models.Task")).Return(repoError)

	// Act
	task, err := service.CreateTask(req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "failed to create task in repository")
	assert.True(t, errors.Is(err, repoError)) // Verify the underlying error is wrapped
	mockRepo.AssertExpectations(t)
}

// --- Test Cases for GetTask ---
func TestGetTask_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	expectedTask := &models.Task{
		ID: 1, Title: "Existing Task", Description: "Desc", Status: "pending",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	// Expect GetByID to be called with ID 1 and return the expected task
	mockRepo.On("GetByID", 1).Return(expectedTask, nil)

	// Act
	task, err := service.GetTask(1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Title, task.Title)
	mockRepo.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	// Expect GetByID to be called with ID 99 and return nil (no task) and an error
	repoError := fmt.Errorf("task with ID %d not found", 99)
	mockRepo.On("GetByID", 99).Return(nil, repoError)

	// Act
	task, err := service.GetTask(99)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "failed to get task from repository")
	assert.True(t, errors.Is(err, repoError))
	mockRepo.AssertExpectations(t)
}

func TestGetTask_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	// Repository method should not be called
	mockRepo.AssertNotCalled(t, "GetByID")

	// Act
	task, err := service.GetTask(0) // Invalid ID

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "invalid task ID", err.Error())
	mockRepo.AssertExpectations(t)
}

// --- Test Cases for GetAllTasks ---
func TestGetAllTasks_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	// Prepare some dummy tasks
	task1 := &models.Task{ID: 1, Title: "Task 1", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	task2 := &models.Task{ID: 2, Title: "Task 2", Status: "completed", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	expectedTasks := []*models.Task{task1, task2}

	// Expect GetAll to be called and return the list of tasks
	mockRepo.On("GetAll").Return(expectedTasks, nil)

	// Act
	tasks, err := service.GetAllTasks()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Len(t, tasks, 2)
	assert.Equal(t, expectedTasks, tasks)
	mockRepo.AssertExpectations(t)
}

func TestGetAllTasks_RepoError(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	repoError := errors.New("failed to fetch from database")
	mockRepo.On("GetAll").Return(nil, repoError)

	// Act
	tasks, err := service.GetAllTasks()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, tasks)
	assert.Contains(t, err.Error(), "failed to get all tasks from repository")
	assert.True(t, errors.Is(err, repoError))
	mockRepo.AssertExpectations(t)
}

// --- Test Cases for UpdateTask ---
func TestUpdateTask_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	existingTask := &models.Task{
		ID: 1, Title: "Original Title", Description: "Original Desc", Status: "pending",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	updateReq := &models.UpdateTaskRequest{
		Title:       "Updated Title",
		Description: "Updated Desc",
		Status:      "in_progress",
	}

	// Mock GetByID to return the existing task
	mockRepo.On("GetByID", 1).Return(existingTask, nil)
	// Mock Update to be called with the modified task and return no error
	mockRepo.On("Update", mock.AnythingOfType("*models.Task")).Return(nil)

	// Act
	updatedTask, err := service.UpdateTask(1, updateReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, "Updated Title", updatedTask.Title)
	assert.Equal(t, "Updated Desc", updatedTask.Description)
	assert.Equal(t, "in_progress", updatedTask.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_PartialUpdate(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	existingTask := &models.Task{
		ID: 1, Title: "Original Title", Description: "Original Desc", Status: "pending",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	updateReq := &models.UpdateTaskRequest{
		Status: "completed", // Only updating status
	}

	mockRepo.On("GetByID", 1).Return(existingTask, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.Task")).Return(nil)

	// Act
	updatedTask, err := service.UpdateTask(1, updateReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, "Original Title", updatedTask.Title)      // Should remain unchanged
	assert.Equal(t, "Original Desc", updatedTask.Description) // Should remain unchanged
	assert.Equal(t, "completed", updatedTask.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	existingTask := &models.Task{ID: 1, Title: "Title", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	updateReq := &models.UpdateTaskRequest{
		Status: "invalid_status", // Invalid status value
	}

	mockRepo.On("GetByID", 1).Return(existingTask, nil)
	// Update should not be called if status validation fails
	mockRepo.AssertNotCalled(t, "Update")

	// Act
	updatedTask, err := service.UpdateTask(1, updateReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, updatedTask)
	assert.Equal(t, "invalid status value", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	updateReq := &models.UpdateTaskRequest{Title: "New Title"}
	repoError := fmt.Errorf("task with ID %d not found: sql: no rows in result set", 99)
	mockRepo.On("GetByID", 99).Return(nil, repoError)

	// Act
	updatedTask, err := service.UpdateTask(99, updateReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, updatedTask)
	assert.Contains(t, err.Error(), "task with ID 99 not found")
	mockRepo.AssertExpectations(t)
}

// --- Test Cases for DeleteTask ---
func TestDeleteTask_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	mockRepo.On("Delete", 1).Return(nil) // Expect Delete with ID 1 to succeed

	// Act
	err := service.DeleteTask(1)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTask_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	repoError := fmt.Errorf("task with ID %d not found for deletion", 99)
	mockRepo.On("Delete", 99).Return(repoError)

	// Act
	err := service.DeleteTask(99)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete task from repository")
	assert.True(t, errors.Is(err, repoError))
	mockRepo.AssertExpectations(t)
}

func TestDeleteTask_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo)

	mockRepo.AssertNotCalled(t, "Delete") // Repository method should not be called

	// Act
	err := service.DeleteTask(0) // Invalid ID

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "invalid task ID", err.Error())
	mockRepo.AssertExpectations(t)
}
