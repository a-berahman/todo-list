// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"context"
)

type Querier interface {
	CreateTodo(ctx context.Context, arg CreateTodoParams) error
}

var _ Querier = (*Queries)(nil)
