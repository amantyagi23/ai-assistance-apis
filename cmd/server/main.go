package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/amantyagi23/ai-assistance/internal/config"
	db "github.com/amantyagi23/ai-assistance/internal/shared/infra/db/mongodb"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/http/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/joho/godotenv"
)

func main() {

	// Load ENV
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	cfg := config.Load()

	db.ConnectMongdb(db.DBConfig{URI: cfg.MongodbURI, Database: cfg.DatabaseName, Timeout: cfg.DatabaseTimeOut})

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {

			code := fiber.StatusInternalServerError
			message := err.Error()

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   message,
			})
		},
		// StrictRouting: true,
		BodyLimit: 10 * 1024 * 1024,
		AppName:   "ai-assistance-api",
		// CaseSensitive: true,
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(helmet.New())

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: cfg.FrontendURL,
	// 	AllowMethods: "GET,POST,PUT,DELETE,PATCH",
	// }))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 60 * 1000000000,
	}))

	// API prefix
	api := app.Group(cfg.APIPrefix)

	routes.RegisterRoutes(api)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	address := "0.0.0.0:" + cfg.Port
	log.Println("Server running on", address)

	log.Fatal(app.Listen(address))
}
