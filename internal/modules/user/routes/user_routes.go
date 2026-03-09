package user

import (
	user "github.com/amantyagi23/ai-assistance/internal/modules/user/controller"
	"github.com/amantyagi23/ai-assistance/internal/shared/infra/middleware"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(route fiber.Router) {
	route.Post("/", user.NewCreateUserController().Execute)
	route.Post("/login", user.NewLoginController().Execute)
	route.Get("/getme", middleware.NewMiddleware().EnsureAuthenticate(), user.NewGetMeController().Execute)
	route.Put("/", middleware.NewMiddleware().EnsureAuthenticate(), user.NewCreateUserController().Execute)
}
