package controller

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// BaseController provides common response methods for all controllers
type BaseController struct{}

// Response standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// MetaInfo contains pagination and other metadata
type MetaInfo struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

// Success sends a success response
func (bc *BaseController) Success(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a created response (201)
func (bc *BaseController) Created(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func (bc *BaseController) Error(c *fiber.Ctx, status int, message string, errCode string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
		},
	})
}

// ErrorWithDetails sends an error response with details
func (bc *BaseController) ErrorWithDetails(c *fiber.Ctx, status int, message string, errCode string, details interface{}) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError sends a validation error response
func (bc *BaseController) ValidationError(c *fiber.Ctx, errors []ValidationError) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errors,
		},
	})
}

// Paginated sends a paginated response
func (bc *BaseController) Paginated(c *fiber.Ctx, data interface{}, page, limit int, total int64) error {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := &MetaInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// NoContent sends a 204 No Content response
func (bc *BaseController) NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// BadRequest sends a 400 Bad Request response
func (bc *BaseController) BadRequest(c *fiber.Ctx, message string) error {
	return bc.Error(c, fiber.StatusBadRequest, message, "BAD_REQUEST")
}

// Unauthorized sends a 401 Unauthorized response
func (bc *BaseController) Unauthorized(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Unauthorized access"
	}
	return bc.Error(c, fiber.StatusUnauthorized, message, "UNAUTHORIZED")
}

// Forbidden sends a 403 Forbidden response
func (bc *BaseController) Forbidden(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Access forbidden"
	}
	return bc.Error(c, fiber.StatusForbidden, message, "FORBIDDEN")
}

// NotFound sends a 404 Not Found response
func (bc *BaseController) NotFound(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return bc.Error(c, fiber.StatusNotFound, message, "NOT_FOUND")
}

// InternalServerError sends a 500 Internal Server Error response
func (bc *BaseController) InternalServerError(c *fiber.Ctx, err error) error {
	message := "Internal server error"
	if err != nil && err.Error() != "" {
		message = err.Error()
	}
	return bc.Error(c, fiber.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR")
}

// Conflict sends a 409 Conflict response
func (bc *BaseController) Conflict(c *fiber.Ctx, message string) error {
	return bc.Error(c, fiber.StatusConflict, message, "CONFLICT")
}

// TooManyRequests sends a 429 Too Many Requests response
func (bc *BaseController) TooManyRequests(c *fiber.Ctx, message string) error {
	return bc.Error(c, fiber.StatusTooManyRequests, message, "TOO_MANY_REQUESTS")
}

// ServiceUnavailable sends a 503 Service Unavailable response
func (bc *BaseController) ServiceUnavailable(c *fiber.Ctx, message string) error {
	return bc.Error(c, fiber.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE")
}

// ValidateRequest validates the request body against a struct
func (bc *BaseController) ValidateRequest(c *fiber.Ctx, request interface{}) (bool, []ValidationError) {
	if err := c.BodyParser(request); err != nil {
		return false, []ValidationError{
			{
				Field:   "body",
				Message: "Invalid request body",
			},
		}
	}

	if err := validate.Struct(request); err != nil {
		var errors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   strings.ToLower(err.Field()),
				Message: bc.getValidationMessage(err),
				Tag:     err.Tag(),
				Value:   err.Param(),
			})
		}
		return false, errors
	}

	return true, nil
}

// ValidateQuery validates query parameters against a struct
func (bc *BaseController) ValidateQuery(c *fiber.Ctx, query interface{}) (bool, []ValidationError) {
	if err := c.QueryParser(query); err != nil {
		return false, []ValidationError{
			{
				Field:   "query",
				Message: "Invalid query parameters",
			},
		}
	}

	if err := validate.Struct(query); err != nil {
		var errors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   strings.ToLower(err.Field()),
				Message: bc.getValidationMessage(err),
				Tag:     err.Tag(),
				Value:   err.Param(),
			})
		}
		return false, errors
	}

	return true, nil
}

// ValidateParams validates URL parameters against a struct
func (bc *BaseController) ValidateParams(c *fiber.Ctx, params interface{}) (bool, []ValidationError) {
	if err := c.ParamsParser(params); err != nil {
		return false, []ValidationError{
			{
				Field:   "params",
				Message: "Invalid URL parameters",
			},
		}
	}

	if err := validate.Struct(params); err != nil {
		var errors []ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   strings.ToLower(err.Field()),
				Message: bc.getValidationMessage(err),
				Tag:     err.Tag(),
				Value:   err.Param(),
			})
		}
		return false, errors
	}

	return true, nil
}

// getValidationMessage returns a human-readable validation message
func (bc *BaseController) getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "gte":
		return "Value must be greater than or equal to " + err.Param()
	case "lte":
		return "Value must be less than or equal to " + err.Param()
	case "len":
		return "Value must be exactly " + err.Param() + " characters long"
	case "numeric":
		return "Value must be numeric"
	case "alpha":
		return "Value must contain only letters"
	case "alphanum":
		return "Value must contain only letters and numbers"
	case "uuid":
		return "Invalid UUID format"
	case "url":
		return "Invalid URL format"
	case "ip":
		return "Invalid IP address format"
	default:
		return "Invalid value"
	}
}
