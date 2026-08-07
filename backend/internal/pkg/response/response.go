package response

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Response is the standard API response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse is the paginated API response structure
type PaginatedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination interface{} `json:"pagination"`
	Summary    interface{} `json:"summary,omitempty"`
}

// Success sends a success response
func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Created sends a created response
func Created(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "created",
		Data:    data,
	})
}

// Error sends an error response
func Error(c *app.RequestContext, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Code:    statusCode,
		Message: message,
	})
}

// Paginated sends a paginated response
func Paginated(c *app.RequestContext, data interface{}, pagination interface{}) {
	c.JSON(consts.StatusOK, PaginatedResponse{
		Code:       http.StatusOK,
		Message:    "success",
		Data:       data,
		Pagination: pagination,
	})
}

// PaginatedWithSummary sends a paginated response with summary
func PaginatedWithSummary(c *app.RequestContext, data interface{}, pagination interface{}, summary interface{}) {
	c.JSON(consts.StatusOK, PaginatedResponse{
		Code:       http.StatusOK,
		Message:    "success",
		Data:       data,
		Pagination: pagination,
		Summary:    summary,
	})
}

// BadRequest sends a bad request response
func BadRequest(c *app.RequestContext, message string) {
	Error(c, consts.StatusBadRequest, message)
}

// Unauthorized sends an unauthorized response
func Unauthorized(c *app.RequestContext, message string) {
	Error(c, consts.StatusUnauthorized, message)
}

// Forbidden sends a forbidden response
func Forbidden(c *app.RequestContext, message string) {
	Error(c, consts.StatusForbidden, message)
}

// NotFound sends a not found response
func NotFound(c *app.RequestContext, message string) {
	Error(c, consts.StatusNotFound, message)
}

// Conflict sends a conflict response
func Conflict(c *app.RequestContext, message string) {
	Error(c, consts.StatusConflict, message)
}

// InternalServerError sends an internal server error response
func InternalServerError(c *app.RequestContext, message string) {
	Error(c, consts.StatusInternalServerError, message)
}

// ValidationError sends a validation error response
func ValidationError(c *app.RequestContext, errors []ValidationErrorDetail) {
	c.JSON(consts.StatusBadRequest, map[string]interface{}{
		"code":    consts.StatusBadRequest,
		"message": "validation failed",
		"errors":  errors,
	})
}

// ValidationErrorDetail represents a single validation error
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
