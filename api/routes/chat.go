package routes

import (
	"github.com/aramceballos/chat-group-server/api/handlers"
	"github.com/aramceballos/chat-group-server/pkg/chat"
	"github.com/gofiber/fiber/v2"
)

func ChatRouter(app fiber.Router, service chat.Service) {
	app.Get("/chat/:channelId", handlers.ChatHandler(service))
}
