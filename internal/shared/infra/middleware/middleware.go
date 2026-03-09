package middleware

import (
	"context"
	"time"

	"github.com/amantyagi23/ai-assistance/internal/config"
	user "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware handles authentication and authorization
type Middleware struct {
	userRepo        *user.UserRepo
	userSessionRepo *user.UserSessionRepo
}

// NewAuthMiddleware creates a new authentication middleware
func NewMiddleware() *Middleware {
	return &Middleware{
		userRepo:        user.NewUserRepo(),
		userSessionRepo: user.NewUserSessionRepo(),
	}
}

func (m *Middleware) EnsureAuthenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Extract token from request
		tokenString, err := m.extractToken(c)
		if err != nil {
			return m.handleError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		}

		// Validate token
		validationResult, err := m.validateToken(tokenString)
		if err != nil {
			return m.handleError(c, fiber.StatusUnauthorized, "INVALID_TOKEN", err.Error())
		}

		// Check session in database
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		session, err := m.validateSession(ctx, validationResult.UserID)
		if err != nil {
			return m.handleError(c, fiber.StatusUnauthorized, "SESSION_ERROR", err.Error())
		}

		// Update session last activity
		go m.updateSessionActivity(context.Background(), session.SessionID.Hex())

		// Set user context
		userCtx := &UserContext{
			UserID: session.UserID,
		}
		SetUserContext(c, userCtx)

		return c.Next()
	}
}

// ValidateToken validates and returns claims from token
func (m *Middleware) validateToken(tokenString string) (*TokenValidationResult, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(config.Load().JWTSecret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &TokenValidationResult{

			UserID: claims.UserID,
		}, nil
	}

	return nil, ErrInvalidToken
}

// validateSession checks if session exists and is valid
func (m *Middleware) validateSession(ctx context.Context, userID string) (*user.UserSession, error) {

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	session, err := m.userSessionRepo.FindLatestActiveSessionByUserID(ctx, userObjID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// updateSessionActivity updates session last activity timestamp
func (m *Middleware) updateSessionActivity(ctx context.Context, sessionID string) {
	_ = m.userSessionRepo.UpdateSessionActivity(ctx, sessionID)
}

// extractToken extracts token from request
func (m *Middleware) extractToken(c *fiber.Ctx) (string, error) {
	// Check Authorization header
	// authHeader := c.Get(config.Load().TokenHeader)
	// if authHeader != "" {
	// 	// Check if it's Bearer token
	// 	parts := strings.Split(authHeader, " ")
	// 	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
	// 		return parts[1], nil
	// 	}
	// 	// If it's just the token without Bearer prefix
	// 	if len(parts) == 1 {
	// 		return parts[0], nil
	// 	}
	// }

	// // Check query parameter
	// token := c.Query("token")
	// if token != "" {
	// 	return token, nil
	// }

	// Check cookie
	token := c.Cookies("access_token")
	if token != "" {
		return token, nil
	}

	return "", ErrNoToken
}

// handleError sends error response
func (m *Middleware) handleError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}
