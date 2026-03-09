package middleware

import "github.com/gofiber/fiber/v2"

func JSONOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Allow GET requests without body
		if c.Method() == fiber.MethodGet {
			return c.Next()
		}

		if c.Get("Content-Type") != "application/json" {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"success": false,
				"error":   "Content-Type must be application/json",
			})
		}

		return c.Next()
	}
}
