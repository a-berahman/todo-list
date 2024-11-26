package schemas

type APIResponse struct {
	Success bool           `json:"success"`         // True for success, false for errors
	Data    interface{}    `json:"data,omitempty"`  // Response data for success (e.g., TodoResponse)
	Error   *ErrorResponse `json:"error,omitempty"` // Error details for failure
}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type TodoResponse struct {
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
	FileID      string `json:"fileId,omitempty"`
}
