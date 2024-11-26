package todo

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-berahman/todo-list/internal/application"
	"github.com/a-berahman/todo-list/internal/domain"
	"github.com/a-berahman/todo-list/internal/handlers/schemas"
	"github.com/a-berahman/todo-list/internal/ports/inbound"

	_ "github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	todoService inbound.TodoService
	logger      *slog.Logger
}

func NewTodoHandler(todoService *application.TodoService, logger *slog.Logger) *TodoHandler {
	return &TodoHandler{todoService: todoService, logger: logger}
}

func (h *TodoHandler) CreateTodo(c echo.Context) error {
	var req schemas.CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "BadRequest",
			Message: "Failed to parse form data",
			Details: err.Error(),
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "ValidationError",
			Message: "Validation failed for one or more fields",
			Details: err.Error(),
		})
	}

	var fileData []byte
	file, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "InvalidFile",
			Message: "Failed to process file",
			Details: err.Error(),
		})
	}

	if file != nil {
		// Validate file extension
		allowedExtensions := map[string]bool{
			".txt": true,
			".png": true,
			".jpg": true,
		}
		filename := file.Filename
		extension := strings.ToLower(filename[strings.LastIndex(filename, "."):])

		isAllowed := false
		if _, ok := allowedExtensions[extension]; ok {
			isAllowed = true
		}

		if !isAllowed {
			return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
				Error:   "InvalidFileType",
				Message: "The uploaded file type is not supported.",
				Details: map[string]interface{}{
					"allowedFileTypes": allowedExtensions,
					"providedFileType": extension,
				},
			})
		}

		// Read file data
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
				Error:   "FailedToOpenFile",
				Message: "Failed to open the uploaded file",
				Details: err.Error(),
			})
		}
		defer src.Close()

		fileData, err = io.ReadAll(src)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
				Error:   "FailedToReadFile",
				Message: "Failed to read the uploaded file",
				Details: err.Error(),
			})
		}
	}

	fileID := uuid.New().String()
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "InvalidDueDate",
			Message: "Invalid due date format",
			Details: err.Error(),
		})
	}
	if err := h.todoService.CreateTodo(c.Request().Context(),
		domain.TodoItem{
			ID:          uuid.New().String(),
			Description: req.Description,
			DueDate:     dueDate,
		}, fileData); err != nil {
		return c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "CreateTodoFailed",
			Message: "Failed to create todo item",
			Details: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, schemas.APIResponse{
		Success: true,
		Data: schemas.TodoResponse{
			Description: req.Description,
			DueDate:     req.DueDate,
			FileID:      fileID,
		},
	})
}
