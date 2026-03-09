package user

import (
	"context"
	"time"

	userDTO "github.com/amantyagi23/ai-assistance/internal/modules/user/dto"
	userRepo "github.com/amantyagi23/ai-assistance/internal/modules/user/repo"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/controller"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

// UpdateUserRequest represents the request body for updating a user

// GetUserParams represents URL parameters for user operations

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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uc.InternalServerError(c, err)
	}

	// Create user
	user := &userRepo.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return uc.InternalServerError(c, err)
	}

	// Remove password from response
	user.Password = ""

	return uc.Created(c, user, "User created successfully")
}

// // UpdateUser handles updating a user
// func (uc *CreateUserController) UpdateUser(c *fiber.Ctx) error {
// 	var params GetUserParams

// 	// Validate params
// 	if valid, errors := uc.ValidateParams(c, &params); !valid {
// 		return uc.ValidationError(c, errors)
// 	}

// 	var req UpdateUserRequest

// 	// Validate request
// 	if valid, errors := uc.ValidateRequest(c, &req); !valid {
// 		return uc.ValidationError(c, errors)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	// Check if user exists
// 	existingUser, err := uc.userRepo.FindByID(ctx, params.ID)
// 	if err != nil {
// 		if err == repository.ErrUserNotFound {
// 			return uc.NotFound(c, "User not found")
// 		}
// 		return uc.InternalServerError(c, err)
// 	}

// 	// Prepare updates
// 	updates := &repository.UpdateUserRequest{
// 		Name:        req.Name,
// 		PhoneNumber: req.PhoneNumber,
// 		Preferences: req.Preferences,
// 	}

// 	if err := uc.userRepo.Update(ctx, params.ID, updates); err != nil {
// 		return uc.InternalServerError(c, err)
// 	}

// 	// Get updated user
// 	updatedUser, err := uc.userRepo.FindByID(ctx, params.ID)
// 	if err != nil {
// 		return uc.InternalServerError(c, err)
// 	}

// 	// Remove password from response
// 	updatedUser.Password = ""

// 	return uc.Success(c, updatedUser, "User updated successfully")
// }

// // DeleteUser handles soft deleting a user
// func (uc *CreateUserController) DeleteUser(c *fiber.Ctx) error {
// 	var params GetUserParams

// 	// Validate params
// 	if valid, errors := uc.ValidateParams(c, &params); !valid {
// 		return uc.ValidationError(c, errors)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	if err := uc.userRepo.Delete(ctx, params.ID); err != nil {
// 		if err == repository.ErrUserNotFound {
// 			return uc.NotFound(c, "User not found")
// 		}
// 		return uc.InternalServerError(c, err)
// 	}

// 	return uc.Success(c, nil, "User deleted successfully")
// }

// // UploadProfilePicture handles uploading user profile picture
// func (uc *CreateUserController) UploadProfilePicture(c *fiber.Ctx) error {
// 	var params GetUserParams

// 	// Validate params
// 	if valid, errors := uc.ValidateParams(c, &params); !valid {
// 		return uc.ValidationError(c, errors)
// 	}

// 	// Get uploaded file
// 	file, err := uc.GetUploadedFile(c, "profile_pic")
// 	if err != nil {
// 		return uc.BadRequest(c, "No file uploaded")
// 	}

// 	// Validate file type
// 	allowedTypes := map[string]bool{
// 		"image/jpeg": true,
// 		"image/png":  true,
// 		"image/gif":  true,
// 		"image/webp": true,
// 	}

// 	if !allowedTypes[file.Header.Get("Content-Type")] {
// 		return uc.BadRequest(c, "Invalid file type. Only JPEG, PNG, GIF, and WEBP are allowed")
// 	}

// 	// Validate file size (max 5MB)
// 	if file.Size > 5*1024*1024 {
// 		return uc.BadRequest(c, "File too large. Maximum size is 5MB")
// 	}

// 	// TODO: Save file to storage and get URL
// 	fileURL := "/uploads/" + file.Filename

// 	// Update user with profile picture URL
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	updates := &repository.UpdateUserRequest{
// 		ProfilePic: fileURL,
// 	}

// 	if err := uc.userRepo.Update(ctx, params.ID, updates); err != nil {
// 		return uc.InternalServerError(c, err)
// 	}

// 	return uc.Success(c, fiber.Map{
// 		"profile_pic": fileURL,
// 	}, "Profile picture uploaded successfully")
// }

// // GetUserStats handles retrieving user statistics
// func (uc *CreateUserController) GetUserStats(c *fiber.Ctx) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	totalUsers, err := uc.userRepo.Count(ctx)
// 	if err != nil {
// 		return uc.InternalServerError(c, err)
// 	}

// 	// Get active users in last 24 hours
// 	// This would require additional repository method

// 	return uc.Success(c, fiber.Map{
// 		"total_users":    totalUsers,
// 		"active_today":   0, // Placeholder
// 		"verified_users": 0, // Placeholder
// 	}, "User statistics retrieved successfully")
// }
