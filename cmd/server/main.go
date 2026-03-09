package main

import (
	"log"

	"github.com/amantyagi23/ai-assistance/internal/config"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, provide system environment variables")
		return
	}

	app := fiber.New()

	api := app.Group(config.APPConfig().APIPrefix)
	routes.RegisterRoutes(api)

	app.Listen("0.0.0.0:" + config.APPConfig().Port)
}
