package chat

import (
	chatDTO "github.com/amantyagi23/ai-assistance/internal/modules/chat/dto"
	chatService "github.com/amantyagi23/ai-assistance/internal/modules/chat/service"
	"github.com/gofiber/fiber/v2"
)

func CreateChat(ctx *fiber.Ctx) error {

	var body chatDTO.CreateChatDTO

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request",
			"success": false,
		})
	}

	// Call service layer
	stream := false
	result, err := chatService.OllamaService(chatService.GenerateRequest{
		Model:  "llama3",
		Prompt: body.Message,
		Stream: &stream,
	})

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate response",
			"success": false,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"reply":   result.Response,
		"success": true,
	})
}
