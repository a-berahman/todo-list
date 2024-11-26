package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCustomValidator(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedError bool
	}{
		{
			name: "valid struct",
			input: struct {
				Field string `validate:"required"`
			}{
				Field: "value",
			},
			expectedError: false,
		},
		{
			name: "invalid struct",
			input: struct {
				Field string `validate:"required"`
			}{
				Field: "",
			},
			expectedError: true,
		},
		{
			name: "valid datetime",
			input: struct {
				Date string `validate:"datetime=2006-01-02T15:04:05Z"`
			}{
				Date: "2024-01-01T12:00:00Z",
			},
			expectedError: false,
		},
		{
			name: "invalid datetime",
			input: struct {
				Date string `validate:"datetime=2006-01-02T15:04:05Z"`
			}{
				Date: "invalid-date",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			v.RegisterValidation("datetime", func(fl validator.FieldLevel) bool {
				_, err := time.Parse(fl.Param(), fl.Field().String())
				return err == nil
			})

			cv := &CustomValidator{Validator: v}
			err := cv.Validate(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequestTimer(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		handlerLatency time.Duration
		expectedStatus int
	}{
		{
			name:           "successful request",
			method:         http.MethodGet,
			path:           "/test",
			handlerLatency: 100 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error request",
			method:         http.MethodPost,
			path:           "/error",
			handlerLatency: 50 * time.Millisecond,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			e := echo.New()
			e.Use(requestTimer(logger))
			e.Add(tt.method, tt.path, func(c echo.Context) error {
				time.Sleep(tt.handlerLatency) // Simulate processing time
				if tt.expectedStatus == http.StatusInternalServerError {
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				return c.String(http.StatusOK, "success")
			})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestGetProjectRoot(t *testing.T) {
	root := getProjectRoot()
	assert.NotEmpty(t, root)

	_, err := os.Stat(root)
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, "go.mod"))
	assert.NoError(t, err)
}
