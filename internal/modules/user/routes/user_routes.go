package user

import (
	user "github.com/amantyagi23/ai-assistance/internal/modules/user/controller"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(route fiber.Router) {
	route.Post("/", user.CreateUser)
}
