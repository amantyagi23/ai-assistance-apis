package routes

import (
	chat "github.com/amantyagi23/ai-assistance/internal/modules/chat/routes"
	user "github.com/amantyagi23/ai-assistance/internal/modules/user/routes"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(routes fiber.Router) {

	// health routes
	routes.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to AI-Assistance Apis",
			"success": true,
		})
	})

	// user routes
	userRoutes := routes.Group("/user")
	user.UserRoutes(userRoutes)

	// chat routes
	chatRoutes := routes.Group("/chat")
	chat.ChatRoutes(chatRoutes)
}
