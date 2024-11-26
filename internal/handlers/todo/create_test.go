package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/a-berahman/todo-list/internal/domain"
	"github.com/a-berahman/todo-list/internal/handlers/schemas"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTodoService struct {
	mock.Mock
}

func (m *MockTodoService) CreateTodo(ctx context.Context, todo domain.TodoItem, fileData []byte) error {
	args := m.Called(ctx, todo, fileData)
	return args.Error(0)
}

type CustomValidator struct{}

func (cv *CustomValidator) Validate(i interface{}) error {
	if i == nil {
		return errors.New("validation failed")
	}
	return nil
}

func TestCreateTodo(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		dueDate        string
		fileContent    []byte
		fileName       string
		fileExtension  string
		setupMock      func(*MockTodoService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "successful creation without file",
			description: "Test todo",
			dueDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			setupMock: func(m *MockTodoService) {
				m.On("CreateTodo", mock.Anything, mock.AnythingOfType("domain.TodoItem"), mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:          "successful creation with valid file",
			description:   "Test todo with file",
			dueDate:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			fileContent:   []byte("test content"),
			fileName:      "test",
			fileExtension: ".txt",
			setupMock: func(m *MockTodoService) {
				m.On("CreateTodo", mock.Anything, mock.AnythingOfType("domain.TodoItem"), mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:          "invalid file extension",
			description:   "Test todo",
			dueDate:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			fileContent:   []byte("test content"),
			fileName:      "test",
			fileExtension: ".invalid",
			setupMock: func(_ *MockTodoService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "invalid due date",
			description: "Test todo",
			dueDate:     "invalid-date",
			setupMock: func(_ *MockTodoService) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:        "service error",
			description: "Test todo",
			dueDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			setupMock: func(m *MockTodoService) {
				m.On("CreateTodo", mock.Anything, mock.AnythingOfType("domain.TodoItem"), mock.Anything).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = &CustomValidator{}

			mockService := &MockTodoService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := &TodoHandler{
				todoService: mockService,
				logger:      slog.Default(),
			}

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("description", tt.description)
			_ = writer.WriteField("dueDate", tt.dueDate)

			if tt.fileContent != nil {
				part, err := writer.CreateFormFile("file", tt.fileName+tt.fileExtension)
				assert.NoError(t, err)
				_, err = part.Write(tt.fileContent)
				assert.NoError(t, err)
			}

			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/todos", body)
			req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.CreateTodo(c)
			assert.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				var errResp schemas.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.NotEmpty(t, errResp.Error)
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				var resp schemas.APIResponse
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.True(t, resp.Success)

				todoResp, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.description, todoResp["description"])
				assert.Equal(t, tt.dueDate, todoResp["dueDate"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
