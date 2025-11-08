package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo represents error information in API response
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

// Error returns an error response
func Error(c *gin.Context, statusCode int, errCode int, message string, details string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
			Details: details,
		},
		RequestID: c.GetString("request_id"),
	})
}

// BadRequest returns a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 400, message, "")
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 401, message, "")
}

// Forbidden returns a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 403, message, "")
}

// NotFound returns a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 404, message, "")
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 500, message, "")
}

// ValidationError returns a 422 Unprocessable Entity response for validation errors
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, 422, message, "")
}