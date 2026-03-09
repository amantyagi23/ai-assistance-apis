package user

import "github.com/gofiber/fiber/v2"

func CreateUser(ctx *fiber.Ctx) error {

	return ctx.SendStatus(201)
}
