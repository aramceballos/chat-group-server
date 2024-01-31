package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aramceballos/chat-group-server/api/routes"
	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "mysecretpassword"
	dbName     = "postgres"
)

// Database instance
var db *sql.DB

func Connect() error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	channelRepo := channel.NewRepository(db)
	channelService := channel.NewService(channelRepo)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	routes.UserRouter(v1, userService)

	routes.ChannelRouter(v1, channelService)

	app.Listen(":4000")
}
