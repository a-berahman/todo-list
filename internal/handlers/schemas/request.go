package schemas

import (
	"mime/multipart"

	_ "github.com/go-playground/validator"
)

type CreateTodoRequest struct {
	Description string                `form:"description" validate:"required,max=255"`
	DueDate     string                `form:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	FileID      *multipart.FileHeader `form:"file" validate:"omitempty"`
}
