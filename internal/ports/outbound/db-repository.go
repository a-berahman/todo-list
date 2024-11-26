package outbound

import (
	"context"

	"github.com/a-berahman/todo-list/internal/infra/db"
)

type DBRepository interface {
	CreateTodo(ctx context.Context, arg db.CreateTodoParams) error
}
