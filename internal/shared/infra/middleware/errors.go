package middleware

import "errors"

var (
	// Token errors
	ErrNoToken          = errors.New("no token provided")
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrTokenRevoked     = errors.New("token has been revoked")

	// Session errors
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session has expired")
	ErrSessionRevoked   = errors.New("session has been revoked")
	ErrInvalidSessionID = errors.New("invalid session ID")

	// User errors
	ErrUserNotFound         = errors.New("user not found")
	ErrUserNotAuthenticated = errors.New("user not authenticated")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrInvalidUserID        = errors.New("invalid user ID")
	ErrInvalidUserEmail     = errors.New("invalid user email")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ErrInvalidUserContext   = errors.New("invalid user context")

	// Permission errors
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrAccessDenied            = errors.New("access denied")
)

// MiddlewareError represents a structured error for middleware
type MiddlewareError struct {
	Code    string
	Message string
	Err     error
	Status  int
}

func (e *MiddlewareError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewMiddlewareError creates a new middleware error
func NewMiddlewareError(code, message string, err error, status int) *MiddlewareError {
	return &MiddlewareError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  status,
	}
}
