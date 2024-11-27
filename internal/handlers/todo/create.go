package todo

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
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

// CreateTodo handles the creation of a new todo item.
func (h *TodoHandler) CreateTodo(c echo.Context) error {
	req, err := h.parseAndValidateRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "BadRequest",
			Message: http.StatusText(http.StatusBadRequest),
			Details: err.Error(),
		})
	}

	fileData, fileID, err := h.processFile(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "FileProcessingError",
			Message: http.StatusText(http.StatusBadRequest),
			Details: err.Error(),
		})
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "InvalidDueDate",
			Message: http.StatusText(http.StatusBadRequest),
			Details: err.Error(),
		})
	}

	todoItem := domain.TodoItem{
		ID:          uuid.New().String(),
		Description: req.Description,
		DueDate:     dueDate,
	}

	if err := h.todoService.CreateTodo(c.Request().Context(), todoItem, fileData); err != nil {
		return c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "CreateTodoFailed",
			Message: http.StatusText(http.StatusInternalServerError),
			Details: err.Error(),
		})
	}

	response := schemas.APIResponse{
		Success: true,
		Data: schemas.TodoResponse{
			Description: req.Description,
			DueDate:     req.DueDate,
			FileID:      fileID,
		},
	}
	return c.JSON(http.StatusCreated, response)
}

func (h *TodoHandler) parseAndValidateRequest(c echo.Context) (*schemas.CreateTodoRequest, error) {
	var req schemas.CreateTodoRequest

	if err := c.Bind(&req); err != nil {
		return nil, errors.New("failed to parse form data")
	}

	if err := c.Validate(req); err != nil {
		return nil, errors.New("validation failed for one or more fields")
	}

	return &req, nil
}

func (h *TodoHandler) processFile(c echo.Context) ([]byte, string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, "", nil
		}
		return nil, "", errors.New("failed to process file")
	}

	if err := h.validateFileExtension(file.Filename); err != nil {
		return nil, "", err
	}

	src, err := file.Open()
	if err != nil {
		return nil, "", errors.New("failed to open the uploaded file")
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, "", errors.New("failed to read the uploaded file")
	}

	fileID := uuid.New().String()
	return fileData, fileID, nil
}

func (h *TodoHandler) validateFileExtension(filename string) error {
	allowedExtensions := map[string]bool{
		".txt": true,
		".png": true,
		".jpg": true,
	}

	extension := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[extension] {
		return errors.New("the uploaded file type is not supported")
	}
	return nil
}
