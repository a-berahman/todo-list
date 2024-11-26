package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-berahman/todo-list/internal/domain"
	"github.com/a-berahman/todo-list/internal/infra/db"
	"github.com/a-berahman/todo-list/internal/ports/outbound"
	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TodoService struct {
	todoRepository   outbound.DBRepository
	fileStorage      outbound.FileStorage
	messagePublisher outbound.MessagePublisher
	logger           *slog.Logger
}

func NewTodoService(todoRepository outbound.DBRepository, fileStorage outbound.FileStorage, messagePublisher outbound.MessagePublisher, logger *slog.Logger) *TodoService {
	return &TodoService{todoRepository: todoRepository, fileStorage: fileStorage, messagePublisher: messagePublisher, logger: logger}
}

func (s *TodoService) CreateTodo(ctx context.Context, todo domain.TodoItem, fileData []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Second)
	defer cancel()

	if err := todo.Validate(); err != nil {
		s.logger.Error("todo validation failed", "error", err)
		return fmt.Errorf("todo validation failed: %w", err)
	}

	todoUUID, err := uuid.Parse(todo.ID)
	if err != nil {
		return fmt.Errorf("invalid todo ID: %w", err)
	}

	if len(fileData) > 0 { // Handle file upload if file data is provided
		fileKey := generateFileKey(todo.ID)
		fileID, err := s.fileStorage.Upload(ctx, fileKey, fileData)
		if err != nil {
			s.logger.Error("failed to upload file", "error", err)
			return fmt.Errorf("failed to upload file: %w", err)
		}
		todo.FileID = fileID
	}

	now := time.Now().UTC()
	createParams := db.CreateTodoParams{
		ID:          pgtype.UUID{Bytes: todoUUID, Valid: true},
		Description: todo.Description,
		DueDate:     pgtype.Timestamp{Time: todo.DueDate, Valid: true},
		FileID:      pgtype.Text{String: todo.FileID, Valid: true},
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}
	if err := s.todoRepository.CreateTodo(ctx, createParams); err != nil {
		return fmt.Errorf("failed to save todo to repository: %w", err)
	}

	todoEvent := domain.TodoItemCreateEvent{
		ID:          todo.ID,
		Description: todo.Description,
		DueDate:     todo.DueDate,
		FileID:      todo.FileID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.publishTodoEvent(ctx, todoEvent); err != nil {
		s.logger.Warn("failed to publish todo event", "error", err)
	}
	return nil
}

func (s *TodoService) publishTodoEvent(ctx context.Context, event domain.TodoItemCreateEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal todo event: %w", err)
	}
	// TODO: we need to make this configurable
	return retry.Do(
		func() error {
			return s.messagePublisher.Publish(ctx, string(eventJSON))
		},
		retry.Attempts(3),
		retry.Delay(500*time.Millisecond),
		retry.LastErrorOnly(true),
	)
}

func generateFileKey(todoID string) string {
	return fmt.Sprintf("todos/%s/%s", todoID, uuid.New().String())
}
