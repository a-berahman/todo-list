package handlers

import (
	"log/slog"

	"github.com/a-berahman/todo-list/internal/application"
	"github.com/a-berahman/todo-list/internal/handlers/todo"
)

type Handler struct {
	TodoHandler *todo.TodoHandler
}

func NewHandler(todoService *application.TodoService, logger *slog.Logger) *Handler {
	return &Handler{TodoHandler: todo.NewTodoHandler(todoService, logger)}
}
