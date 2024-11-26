// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: todo.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTodo = `-- name: CreateTodo :exec
INSERT INTO todo_items (
    id, description, due_date, file_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id
`

type CreateTodoParams struct {
	ID          pgtype.UUID      `json:"id"`
	Description string           `json:"description"`
	DueDate     pgtype.Timestamp `json:"dueDate"`
	FileID      pgtype.Text      `json:"fileId"`
	CreatedAt   pgtype.Timestamp `json:"createdAt"`
	UpdatedAt   pgtype.Timestamp `json:"updatedAt"`
}

func (q *Queries) CreateTodo(ctx context.Context, arg CreateTodoParams) error {
	_, err := q.db.Exec(ctx, createTodo,
		arg.ID,
		arg.Description,
		arg.DueDate,
		arg.FileID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}
