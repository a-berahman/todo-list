package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/a-berahman/todo-list/internal/domain"
	"github.com/a-berahman/todo-list/internal/infra/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockDBRepository struct {
	mock.Mock
}

func (m *MockDBRepository) CreateTodo(ctx context.Context, arg db.CreateTodoParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) Upload(ctx context.Context, key string, file []byte) (string, error) {
	args := m.Called(ctx, key, file)
	return args.String(0), args.Error(1)
}

type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) Publish(ctx context.Context, message string) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func TestCreateTodo(t *testing.T) {

	validUUID := uuid.New().String()
	futureTime := time.Now().Add(24 * time.Hour)
	fileData := []byte("test file content")

	tests := []struct {
		name          string
		todo          domain.TodoItem
		fileData      []byte
		setupMocks    func(*MockDBRepository, *MockFileStorage, *MockMessagePublisher)
		expectedError string
	}{
		{
			name: "successful creation without file",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "Test todo",
				DueDate:     futureTime,
			},
			fileData: nil,
			setupMocks: func(mockDB *MockDBRepository, fs *MockFileStorage, mp *MockMessagePublisher) {
				mockDB.On("CreateTodo", mock.Anything, mock.MatchedBy(func(params db.CreateTodoParams) bool {
					return params.Description == "Test todo"
				})).Return(nil)
				mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "successful creation with file",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "Test todo with file",
				DueDate:     futureTime,
			},
			fileData: fileData,
			setupMocks: func(mockDB *MockDBRepository, fs *MockFileStorage, mp *MockMessagePublisher) {
				fs.On("Upload", mock.Anything, mock.Anything, fileData).Return("file-id", nil)
				mockDB.On("CreateTodo", mock.Anything, mock.MatchedBy(func(params db.CreateTodoParams) bool {
					return params.Description == "Test todo with file" && params.FileID.String == "file-id"
				})).Return(nil)
				mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "validation error - past due date",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "Test todo",
				DueDate:     time.Now().Add(-24 * time.Hour),
			},
			setupMocks: func(_ *MockDBRepository, _ *MockFileStorage, _ *MockMessagePublisher) {

			},
			expectedError: "todo validation failed: due date must be in the future",
		},
		{
			name: "validation error - empty description",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "",
				DueDate:     futureTime,
			},
			setupMocks: func(_ *MockDBRepository, _ *MockFileStorage, _ *MockMessagePublisher) {
			},
			expectedError: "todo validation failed: description cannot be empty",
		},
		{
			name: "invalid UUID",
			todo: domain.TodoItem{
				ID:          "invalid-uuid",
				Description: "Test todo",
				DueDate:     futureTime,
			},
			setupMocks: func(_ *MockDBRepository, _ *MockFileStorage, _ *MockMessagePublisher) {

			},
			expectedError: "invalid todo ID",
		},
		{
			name: "file upload error",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "Test todo",
				DueDate:     futureTime,
			},
			fileData: fileData,
			setupMocks: func(db *MockDBRepository, fs *MockFileStorage, mp *MockMessagePublisher) {
				fs.On("Upload", mock.Anything, mock.Anything, fileData).Return("", errors.New("upload failed"))
			},
			expectedError: "failed to upload file",
		},
		{
			name: "database error",
			todo: domain.TodoItem{
				ID:          validUUID,
				Description: "Test todo",
				DueDate:     futureTime,
			},
			setupMocks: func(db *MockDBRepository, fs *MockFileStorage, mp *MockMessagePublisher) {
				db.On("CreateTodo", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: "failed to save todo to repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDB := new(MockDBRepository)
			mockFS := new(MockFileStorage)
			mockMP := new(MockMessagePublisher)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockFS, mockMP)
			}

			service := NewTodoService(mockDB, mockFS, mockMP, slog.Default())

			err := service.CreateTodo(context.Background(), tt.todo, tt.fileData)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
			mockFS.AssertExpectations(t)
			mockMP.AssertExpectations(t)
		})
	}
}

func TestPublishTodoEvent(t *testing.T) {
	tests := []struct {
		name          string
		event         domain.TodoItemCreateEvent
		setupMock     func(*MockMessagePublisher)
		expectedError string
	}{
		{
			name: "successful publish",
			event: domain.TodoItemCreateEvent{
				ID:          uuid.New().String(),
				Description: "Test event",
				DueDate:     time.Now().Add(24 * time.Hour),
			},
			setupMock: func(mp *MockMessagePublisher) {
				mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "publish error with retry",
			event: domain.TodoItemCreateEvent{
				ID:          uuid.New().String(),
				Description: "Test event",
			},
			setupMock: func(mp *MockMessagePublisher) {
				mp.On("Publish", mock.Anything, mock.Anything).
					Return(errors.New("publish failed")).Times(3)
			},
			expectedError: "publish failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMP := new(MockMessagePublisher)
			if tt.setupMock != nil {
				tt.setupMock(mockMP)
			}

			service := NewTodoService(nil, nil, mockMP, slog.Default())
			err := service.publishTodoEvent(context.Background(), tt.event)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockMP.AssertExpectations(t)
		})
	}
}

func TestGenerateFileKey(t *testing.T) {
	todoID := uuid.New().String()
	key := generateFileKey(todoID)

	assert.Contains(t, key, todoID)
	assert.Contains(t, key, "todos/")
	assert.Regexp(t, `^todos/[^/]+/[^/]+$`, key)
}
