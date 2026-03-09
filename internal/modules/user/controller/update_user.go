package user

import (
	"context"
	"time"

	userDTO "github.com/amantyagi23/ai-assistance/internal/modules/user/dto"
	userRepo "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/controller"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/middleware"
	"github.com/gofiber/fiber/v2"
)

type UpdateUserController struct {
	controller.BaseController
	userRepo *userRepo.UserRepo
}

func NewUpdateUserController() *UpdateUserController {
	return &UpdateUserController{
		userRepo: userRepo.NewUserRepo(),
	}
}

// UpdateUser handles updating a user
func (uc *UpdateUserController) UpdateUser(c *fiber.Ctx) error {

	id, err := middleware.GetUserIDFromContext(c)

	var req userDTO.UpdateUserRequest

	// Validate request
	if valid, errors := uc.ValidateRequest(c, &req); !valid {
		return uc.ValidationError(c, errors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user exists
	existingUser, err := uc.userRepo.FindByID(ctx, id.Hex())
	if err != nil {
		if err == userRepo.ErrUserNotFound {
			return uc.NotFound(c, "User not found")
		}
		return uc.InternalServerError(c, err)
	}

	// Prepare updates
	updates := &userRepo.UpdateUserRequest{
		Name: req.Name,
	}

	if err := uc.userRepo.Update(ctx, id.Hex(), updates); err != nil {
		return uc.InternalServerError(c, err)
	}

	// Get updated user
	updatedUser, err := uc.userRepo.FindByID(ctx, existingUser.UserID.Hex())
	if err != nil {
		return uc.InternalServerError(c, err)
	}

	// Remove password from response
	updatedUser.Password = ""

	return uc.Success(c, updatedUser, "User updated successfully")
}
