package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/aramceballos/chat-group-server/api/routes"
	"github.com/aramceballos/chat-group-server/pkg/auth"
	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"

	_ "github.com/lib/pq"
)

type Message struct {
	Content string `json:"content"`
}
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Database instance
var db *sql.DB

func Connect(
	dbHost string,
	dbPort string,
	dbUser string,
	dbPassword string,
	dbName string,
) error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

type Client struct {
	isClosing bool
	mu        sync.Mutex
}

var channels = make(map[int64]map[*websocket.Conn]*Client)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if err := Connect(dbHost, dbPort, dbUser, dbPassword, dbName); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	channelRepo := channel.NewRepository(db)
	channelService := channel.NewService(channelRepo)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	routes.UserRouter(v1, userService)
	routes.ChannelRouter(v1, channelService)
	routes.AuthRouter(v1, authService)

	v1.Use("/chat", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	v1.Get("/chat/:channelId", websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		if token == "" {
			errorMessage := Result{
				Success: false,
				Message: "Token is required",
			}
			err := c.WriteJSON(errorMessage)
			if err != nil {
				log.Println("write error:", err)
				return
			}
			return
		}

		// Parse the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key
			return []byte("secret"), nil
		})
		if err != nil {
			errorMessage := Result{
				Success: false,
				Message: err.Error(),
			}
			err = c.WriteJSON(errorMessage)
			if err != nil {
				log.Println("write error:", err)
				return
			}
			return
		}

		userId := parsedToken.Claims.(jwt.MapClaims)["user_id"].(float64)

		channelId := c.Params("channelId")
		// Cast channelId to int64
		channelIdInt, err := strconv.ParseInt(channelId, 10, 64)
		if err != nil {
			log.Println("channelId error:", err)
			return
		}

		// Create a new client
		client := &Client{}

		// Add client to the channel
		if _, ok := channels[channelIdInt]; !ok {
			channels[channelIdInt] = make(map[*websocket.Conn]*Client)
		}
		channels[channelIdInt][c] = client

		// Remove client from the channel
		defer func() {
			delete(channels[channelIdInt], c)
		}()

		for {
			msg := Message{}
			err := c.ReadJSON(&msg)
			if err != nil {
				errorMessage := Result{
					Success: false,
					Message: err.Error(),
				}
				err = c.WriteJSON(errorMessage)
				if err != nil {
					log.Println("write error:", err)
					break
				}
				continue
			}

			// Check if message is empty
			if msg.Content == "" {
				continue
			}

			insertedMessage := entities.Message{}
			// Insert message into database
			err = db.QueryRow("INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, user_id, channel_id, content, created_at", channelId, userId, string(msg.Content)).Scan(&insertedMessage.ID, &insertedMessage.UserID, &insertedMessage.ChannelID, &insertedMessage.Content, &insertedMessage.CreatedAt)
			if err != nil {
				errorMessage := Result{
					Success: false,
					Message: err.Error(),
				}
				err = c.WriteJSON(errorMessage)
				if err != nil {
					log.Println("write error:", err)
					break
				}
				continue
			}

			// Query user from database to populate the message
			user := entities.User{}
			err = db.QueryRow("SELECT id, name, avatar_url, created_at FROM users WHERE id = $1", userId).Scan(&user.ID, &user.Name, &user.AvatarURL, &user.CreatedAt)
			if err != nil {
				errorMessage := Result{
					Success: false,
					Message: err.Error(),
				}
				err = c.WriteJSON(errorMessage)
				if err != nil {
					log.Println("write error:", err)
					break
				}
				continue
			}
			insertedMessage.User = user
			fmt.Println("Message sent successfully", insertedMessage)

			res := Result{
				Success: true,
				Message: "Message sent successfully",
			}
			err = c.WriteJSON(res)
			if err != nil {
				log.Println("write error:", err)
				break
			}

			// Send message to all clients
			for conn, cl := range channels[channelIdInt] {
				go func(conn *websocket.Conn, cl *Client) {
					cl.mu.Lock()
					defer cl.mu.Unlock()
					if cl.isClosing {
						return
					}

					err = conn.WriteJSON(insertedMessage)
					if err != nil {
						cl.isClosing = true
						log.Println("write error:", err)

						conn.WriteMessage(websocket.CloseMessage, []byte{})
						conn.Close()
					}
				}(conn, cl)
			}

		}
	}))

	app.Listen(":4000")
}
