package handlers

import (
	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/gofiber/fiber/v2"
)

func GetChannels(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		channels, err := service.FetchChannels()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   channels,
		})
	}

}

func GetChannelById(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		channelId := c.Params("id")
		channel, err := service.FetchChannelById(channelId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   channel,
		})
	}
}
