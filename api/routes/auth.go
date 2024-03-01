package routes

import (
	"github.com/aramceballos/chat-group-server/api/handlers"
	"github.com/aramceballos/chat-group-server/pkg/auth"
	"github.com/aramceballos/chat-group-server/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func AuthRouter(app fiber.Router, service auth.Service) {
	app.Post("/auth/login", handlers.Login(service))
	app.Post("/auth/signup", handlers.Signup(service))
	app.Get("/auth/me", middleware.Protected(), handlers.Me(service))
	app.Put("/auth/change-password", middleware.Protected(), handlers.ChangePassword(service))
}
