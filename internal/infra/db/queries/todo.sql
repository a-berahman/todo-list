-- name: CreateTodo :exec
INSERT INTO todo_items (
    id, description, due_date, file_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id;
