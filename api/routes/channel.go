package routes

import (
	"github.com/aramceballos/chat-group-server/api/handlers"
	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/aramceballos/chat-group-server/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func ChannelRouter(app fiber.Router, service channel.Service) {
	app.Get("/channels", handlers.GetChannels(service))
	app.Get("/channels/:id", handlers.GetChannelById(service))
	app.Post("/channels/new", handlers.CreateChannel(service))
	app.Post("/channels/join", middleware.Protected(), handlers.JoinChannel(service))
	app.Post("/channels/leave", middleware.Protected(), handlers.LeaveChannel(service))
}
