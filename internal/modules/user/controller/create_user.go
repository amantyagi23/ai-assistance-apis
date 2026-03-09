package user

import (
	"context"
	"time"

	userDTO "github.com/amantyagi23/ai-assistance/internal/modules/user/dto"
	userRepo "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/controller"
	"github.com/amantyagi23/ai-assistance/internal/shared/utils"
	"github.com/gofiber/fiber/v2"
)

type CreateUserController struct {
	controller.BaseController
	userRepo *userRepo.UserRepo
}

func NewCreateUserController() *CreateUserController {
	return &CreateUserController{
		userRepo: userRepo.NewUserRepo(),
	}
}

// CreateUser handles user creation
func (uc *CreateUserController) Execute(c *fiber.Ctx) error {
	var req userDTO.CreateUserRequest

	// Validate request
	if valid, errors := uc.ValidateRequest(c, &req); !valid {
		return uc.ValidationError(c, errors)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	existingUser, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return uc.Conflict(c, "User with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPwd(req.Password)
	if err != nil {
		return uc.InternalServerError(c, err)
	}

	// Create user
	user := &userRepo.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return uc.InternalServerError(c, err)
	}

	// Remove password from response
	user.Password = ""

	return uc.Created(c, user, "User created successfully")
}
