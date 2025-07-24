package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

const MaxMessageLength = 10 * 1024 // 10KB max raw message length

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Message struct {
	Body json.RawMessage `json:"body"`
}

type Client struct {
	conn *websocket.Conn
	send chan interface{}
	quit chan struct{}
	once sync.Once
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan interface{}, 256),
		quit: make(chan struct{}),
	}
}

func validateMessageBody(body map[string]interface{}) error {
	msgType, ok := body["type"].(string)
	if !ok {
		return errors.New("Message body must have a 'type' field")
	}

	switch msgType {
	case "text":
		content, ok := body["content"].(string)
		if !ok || len(content) == 0 {
			return errors.New("text messages must have a non-empty 'content' field")
		}
	case "file":
		// Validate file fields
		fileID, fileIDOk := body["file_id"].(string)
		filename, filenameOk := body["filename"].(string)
		mimeType, mimeTypeOk := body["mime_type"].(string)
		url, urlOk := body["url"].(string)
		_, sizeOk := body["size_in_bytes"]
		if !fileIDOk || fileID == "" || !filenameOk || filename == "" || !mimeTypeOk || mimeType == "" || !urlOk || url == "" || !sizeOk {
			return errors.New("file messages must have non-empty 'file_id', 'filename', 'mime_type', 'url', and 'size_in_bytes' fields")
		}
	default:
		return errors.New("unsupported message type")
	}

	return nil
}

func validateToken(token string) (int, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			panic("JWT_SECRET env var must be set")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims format")
	}

	userId, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in token")
	}

	return int(userId), nil
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message := <-c.send:
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("write error [broadcast]: %v", err)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case <-c.quit:
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.quit)
	})
}

var (
	channels   = make(map[int64]map[*websocket.Conn]*Client)
	channelsMu sync.RWMutex
)

func ChatHandler(db *sql.DB) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		if token == "" {
			errorMessage := Result{
				Success: false,
				Message: "Token is required",
			}
			err := c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		userId, err := validateToken(token)
		if err != nil {
			errorMessage := Result{
				Success: false,
				Message: "Token is required",
			}
			err := c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		channelId := c.Params("channelId")
		// Cast channelId to int64
		channelIdInt, err := strconv.ParseInt(channelId, 10, 64)
		if err != nil {
			log.Println("channelId error:", err)
			return
		}

		// Check if user is a member of the channel
		var exists bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM memberships WHERE channel_id = $1 AND user_id = $2)", channelIdInt, userId).Scan(&exists)
		if err != nil || !exists {
			fmt.Println("error checking membership:", err)
			errorMessage := Result{
				Success: false,
				Message: "You are not a member of this channel",
			}
			err = c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		// Create a new client
		client := NewClient(c)

		// Add client to the channel
		channelsMu.Lock()
		if _, ok := channels[channelIdInt]; !ok {
			channels[channelIdInt] = make(map[*websocket.Conn]*Client)
		}
		channels[channelIdInt][c] = client
		channelsMu.Unlock()

		fmt.Printf("Client %s connected to channel %d\n", c.RemoteAddr().String(), channelIdInt)

		go client.writePump()

		// Remove client from the channel
		defer func() {
			fmt.Printf("Client %s disconnected from channel %d\n", c.RemoteAddr().String(), channelIdInt)
			client.close()

			channelsMu.Lock()
			delete(channels[channelIdInt], c)
			if len(channels[channelIdInt]) == 0 {
				delete(channels, channelIdInt)
				fmt.Printf("Channel %d cleaned up (empty)\n", channelIdInt)
			}
			channelsMu.Unlock()
		}()

		for {
			msg := Message{}
			err := c.ReadJSON(&msg.Body)
			if err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			// Check if message is empty
			if len(msg.Body) == 0 {
				continue
			}

			// Validate message does not exceed max length
			if len(msg.Body) > MaxMessageLength {
				client.send <- Result{
					Success: false,
					Message: "Message size exceeds limit",
				}
				continue
			}

			// Validate message body JSON structure
			var body map[string]interface{}
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				client.send <- Result{
					Success: false,
					Message: "Invalid message body JSON",
				}
				continue
			}

			if err = validateMessageBody(body); err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			fmt.Printf("Received message from %s content: %s\n", c.RemoteAddr().String(), string(msg.Body))

			insertedMessage := entities.Message{}
			// Insert message into database
			err = db.QueryRow("INSERT INTO messages (channel_id, user_id, body) VALUES ($1, $2, $3::jsonb) RETURNING id, user_id, channel_id, body, created_at", channelId, userId, msg.Body).Scan(&insertedMessage.ID, &insertedMessage.UserID, &insertedMessage.ChannelID, &insertedMessage.Body, &insertedMessage.CreatedAt)
			if err != nil {
				fmt.Println("error inserting message:", err)
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			// Query user from database to populate the message
			user := entities.User{}
			err = db.QueryRow("SELECT id, name, avatar_url, created_at FROM users WHERE id = $1", userId).Scan(&user.ID, &user.Name, &user.AvatarURL, &user.CreatedAt)
			if err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}
			insertedMessage.User = user
			fmt.Println("Broadcasting message", string(insertedMessage.Body))

			client.send <- Result{
				Success: true,
				Message: "Message sent successfully",
			}

			// Send message to all clients
			for conn, cl := range channels[channelIdInt] {
				go func(conn *websocket.Conn, cl *Client) {
					fmt.Printf("Sending message to %s content: %s\n", conn.RemoteAddr().String(), string(msg.Body))
					cl.send <- insertedMessage
				}(conn, cl)
			}

		}
	})
}
