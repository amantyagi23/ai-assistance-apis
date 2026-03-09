package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserIDFromContext retrieves user ID from Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (primitive.ObjectID, error) {
	userID := c.Locals(string(UserIDKey))
	if userID == nil {
		return primitive.NilObjectID, ErrUserNotAuthenticated
	}

	switch v := userID.(type) {
	case string:
		return primitive.ObjectIDFromHex(v)
	case primitive.ObjectID:
		return v, nil
	default:
		return primitive.NilObjectID, ErrInvalidUserID
	}
}

// SetUserContext sets user information in context
func SetUserContext(c *fiber.Ctx, userCtx *UserContext) {
	c.Locals(string(UserIDKey), userCtx.UserID)

}

// ClearUserContext clears user information from context
func ClearUserContext(c *fiber.Ctx) {
	c.Locals(string(UserIDKey), nil)

}
