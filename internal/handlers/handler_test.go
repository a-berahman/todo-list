package handlers

import (
	"testing"

	"log/slog"

	"github.com/a-berahman/todo-list/internal/application"
	"github.com/a-berahman/todo-list/internal/handlers/todo"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name        string
		todoService *application.TodoService
		logger      *slog.Logger
		want        *Handler
	}{
		{
			name:        "should create new handler successfully",
			todoService: &application.TodoService{},
			logger:      slog.Default(),
			want: &Handler{
				TodoHandler: todo.NewTodoHandler(&application.TodoService{}, slog.Default()),
			},
		},
		{
			name:        "should handle nil service",
			todoService: nil,
			logger:      slog.Default(),
			want: &Handler{
				TodoHandler: todo.NewTodoHandler(nil, slog.Default()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHandler(tt.todoService, tt.logger)
			assert.NotNil(t, got)
			assert.IsType(t, tt.want, got)
			assert.NotNil(t, got.TodoHandler)
		})
	}
}
