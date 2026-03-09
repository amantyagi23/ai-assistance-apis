package chat

import (
	chat "github.com/amantyagi23/ai-assistance/internal/modules/chat/controller"
	"github.com/gofiber/fiber/v2"
)

func ChatRoutes(route fiber.Router) {
	route.Post("/", chat.CreateChat)
}
