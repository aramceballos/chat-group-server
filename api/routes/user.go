package routes

import (
	"github.com/aramceballos/chat-group-server/api/handlers"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/fiber/v2"
)

func UserRouter(app fiber.Router, service user.Service) {
	app.Get("/users", handlers.GetUsers(service))
	app.Get("/users/:id", handlers.GetUserById(service))
}
