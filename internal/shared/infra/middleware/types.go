package middleware

import (
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context keys for storing values in Fiber context
const (
	UserIDKey ContextKey = "userID"
)

// TokenType represents the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents custom JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	UserID string
}

// UserContext represents the user information stored in context
type UserContext struct {
	UserID primitive.ObjectID `json:"id"`
}
