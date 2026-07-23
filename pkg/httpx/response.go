package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApiResponse is the standard JSON envelope returned by every endpoint, so
// clients can rely on a single shape for both success and error responses:
//
//	success: {"status": 200, "data": {...}}
//	error:   {"status": 404, "error": "user not found"}
//
// It is built with the fluent With* methods and written with Send. Optional
// fields are omitted when empty to keep bodies small.
type ApiResponse struct {
	Status  int          `json:"status"`
	Data    any          `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
	Details []FieldError `json:"details,omitempty"`
	Page    int          `json:"page,omitempty"`
	Total   int          `json:"total,omitempty"`
}

// NewResponse starts a response builder with the given HTTP status code.
func NewResponse(status int) ApiResponse {
	return ApiResponse{Status: status}
}

func (r ApiResponse) WithStatus(status int) ApiResponse {
	r.Status = status
	return r
}

func (r ApiResponse) WithData(data any) ApiResponse {
	r.Data = data
	return r
}

func (r ApiResponse) WithError(err string) ApiResponse {
	r.Error = err
	return r
}

// WithDetails attaches field-level validation failures to the response.
func (r ApiResponse) WithDetails(fields []FieldError) ApiResponse {
	r.Details = fields
	return r
}

func (r ApiResponse) WithPage(page, total int) ApiResponse {
	r.Page = page
	r.Total = total
	return r
}

// Send writes the response as JSON, using its Status as the HTTP status code.
// Callers should return immediately after invoking it.
func (r ApiResponse) Send(c *gin.Context) {
	c.JSON(r.Status, r)
}

// WriteSuccess writes a success response carrying data under the given status.
func WriteSuccess(c *gin.Context, status int, data any) {
	NewResponse(status).WithData(data).Send(c)
}

// WriteError writes an error response with the given status and message.
func WriteError(c *gin.Context, status int, message string) {
	NewResponse(status).WithError(message).Send(c)
}

// WriteValidationError writes a 400 whose details list the failed fields.
func WriteValidationError(c *gin.Context, fields []FieldError) {
	NewResponse(http.StatusBadRequest).
		WithError("validation failed").
		WithDetails(fields).
		Send(c)
}
