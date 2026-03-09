package user

import (
	"context"
	"fmt"
	"time"

	"github.com/amantyagi23/ai-assistance/internal/config"
	userDTO "github.com/amantyagi23/ai-assistance/internal/modules/user/dto"
	userRepo "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/controller"
	"github.com/amantyagi23/ai-assistance/internal/shared/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	User         *userRepo.User `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
}

type LoginController struct {
	controller.BaseController
	userRepo        *userRepo.UserRepo
	userSessionRepo *userRepo.UserSessionRepo
}

func NewLoginController() *LoginController {
	return &LoginController{
		userRepo:        userRepo.NewUserRepo(),
		userSessionRepo: userRepo.NewUserSessionRepo(),
	}
}

// CreateUser handles user creation
func (uc *LoginController) Execute(c *fiber.Ctx) error {
	var req userDTO.LoginUserDTO

	// Validate request
	if valid, errors := uc.ValidateRequest(c, &req); !valid {
		return uc.ValidationError(c, errors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
		})
	}
	if user == nil {
		return uc.NotFound(c, "User with this email not exists")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate tokens
	accessToken, refreshToken, err := uc.generateTokens(user.UserID.Hex())
	fmt.Printf(accessToken, refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Create session
	session := &userRepo.UserSession{
		UserID:       user.UserID,
		SessionToken: accessToken,
		RefreshToken: refreshToken,
		SessionType:  userRepo.SessionTypeLogin,
		IPAddress:    c.IP(),
		UserAgent:    c.Get("User-Agent"),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		DeviceInfo: map[string]interface{}{
			"type": c.Get("Device-Type", "unknown"),
		},
	}

	if err := uc.userSessionRepo.CreateSession(ctx, session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}
	expiration := time.Now().Add(24 * time.Hour)

	utils.SetCookieHandler(c, "access_token", accessToken, expiration)
	utils.SetCookieHandler(c, "refresh_token", refreshToken, expiration)

	// Update last login
	_ = uc.userRepo.UpdateLastLogin(ctx, user.UserID.Hex())

	return c.JSON(LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
	})
}

func (h *LoginController) Logout(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find and revoke session
	session, err := h.userSessionRepo.FindBySessionToken(ctx, token)
	if err == nil {
		_ = h.userSessionRepo.RevokeSession(ctx, session.SessionID.Hex())
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func (h *LoginController) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find session by refresh token
	session, err := h.userSessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	// Generate new tokens
	accessToken, refreshToken, err := h.generateTokens(session.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate tokens",
		})
	}

	// Revoke old session
	_ = h.userSessionRepo.RevokeSession(ctx, session.SessionID.Hex())

	// Create new session
	newSession := &userRepo.UserSession{
		UserID:       session.UserID,
		SessionToken: accessToken,
		RefreshToken: refreshToken,
		SessionType:  userRepo.SessionTypeRefresh,
		IPAddress:    c.IP(),
		UserAgent:    c.Get("User-Agent"),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if err := h.userSessionRepo.CreateSession(ctx, newSession); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    24 * 60 * 60,
	})
}

func (h *LoginController) GetActiveSessions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessions, err := h.userSessionRepo.GetUserActiveSessions(ctx, objID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sessions",
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}

func (h *LoginController) RevokeSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	userID := c.Locals("userID").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify session belongs to user
	session, err := h.userSessionRepo.FindBySessionToken(ctx, sessionID)
	if err == nil && session.UserID.Hex() != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot revoke this session",
		})
	}

	if err := h.userSessionRepo.RevokeSession(ctx, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session revoked successfully",
	})
}

// Helper methods
func (h *LoginController) generateTokens(userID string) (string, string, error) {
	// Generate access token
	secretBytes := []byte(config.Load().JWTSecret)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString(secretBytes)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString(secretBytes)
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}
