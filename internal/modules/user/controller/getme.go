package user

import (
	"context"
	"time"

	userRepo "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/controller"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/middleware"
	"github.com/gofiber/fiber/v2"
)

type GetMeController struct {
	controller.BaseController
	userRepo        *userRepo.UserRepo
	userSessionRepo *userRepo.UserSessionRepo
}

func NewGetMeController() *GetMeController {
	return &GetMeController{
		userRepo:        userRepo.NewUserRepo(),
		userSessionRepo: userRepo.NewUserSessionRepo(),
	}
}

// CreateUser handles user creation
func (this *GetMeController) Execute(c *fiber.Ctx) error {

	// Validate request

	id, err := middleware.GetUserIDFromContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	user, err := this.userRepo.FindByID(ctx, id.Hex())
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
		})
	}
	if user == nil {
		return this.NotFound(c, "User Not Found")
	}

	return this.Success(c, user, "")
}
