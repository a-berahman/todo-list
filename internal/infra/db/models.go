// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type TodoItem struct {
	ID          pgtype.UUID      `json:"id"`
	Description string           `json:"description"`
	DueDate     pgtype.Timestamp `json:"dueDate"`
	FileID      pgtype.Text      `json:"fileId"`
	CreatedAt   pgtype.Timestamp `json:"createdAt"`
	UpdatedAt   pgtype.Timestamp `json:"updatedAt"`
}