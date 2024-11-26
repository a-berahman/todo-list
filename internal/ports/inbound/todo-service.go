package inbound

import (
	"context"

	"github.com/a-berahman/todo-list/internal/domain"
)

type TodoService interface {
	CreateTodo(ctx context.Context, todo domain.TodoItem, fileData []byte) error
}
