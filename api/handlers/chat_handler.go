package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aramceballos/chat-group-server/pkg/chat"
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
	conn         *websocket.Conn
	send         chan interface{}
	quit         chan struct{}
	once         sync.Once
	userId       int
	channelId    int
	failureCount int
	lastFailure  time.Time
}

func NewClient(conn *websocket.Conn, userId int, channelId int) *Client {
	return &Client{
		conn:         conn,
		send:         make(chan interface{}, 256),
		quit:         make(chan struct{}),
		userId:       userId,
		channelId:    channelId,
		failureCount: 0,
		lastFailure:  time.Now(),
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

func sendWSError(conn *websocket.Conn, message string) {
	errorMessage := Result{
		Success: false,
		Message: message,
	}
	err := conn.WriteJSON(errorMessage)
	if err != nil {
		log.Printf("write error [error response]: %v", err)
	}
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

type ChannelsHub struct {
	channels   map[int64]map[*websocket.Conn]*Client
	channelsMu sync.RWMutex
}

func NewChannelsHub() *ChannelsHub {
	return &ChannelsHub{
		channels: make(map[int64]map[*websocket.Conn]*Client),
	}
}

func (ch *ChannelsHub) AddClient(channelId int64, conn *websocket.Conn, client *Client) {
	ch.channelsMu.Lock()
	defer ch.channelsMu.Unlock()
	// Create client map for channel if doesn't exist
	if _, ok := ch.channels[channelId]; !ok {
		ch.channels[channelId] = make(map[*websocket.Conn]*Client)
	}
	ch.channels[channelId][conn] = client
}

func (ch *ChannelsHub) RemoveClient(channelId int64, conn *websocket.Conn) {
	ch.channelsMu.Lock()
	defer ch.channelsMu.Unlock()
	delete(ch.channels[channelId], conn)
	// Clean up channel if empty
	if len(ch.channels[channelId]) == 0 {
		delete(ch.channels, channelId)
		log.Printf("Channel %d cleaned up (empty channel)", channelId)
	}
}

func (ch *ChannelsHub) BroadcastMessage(channelId int64, message entities.Message) {
	ch.channelsMu.RLock()
	clients := make([]*Client, 0, len(ch.channels[channelId]))
	for _, client := range ch.channels[channelId] {
		clients = append(clients, client)
	}
	ch.channelsMu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
			client.failureCount = 0
		default:
			log.Printf("Message dropped for client %d on channel %d", client.userId, client.channelId)
			client.failureCount++
			client.lastFailure = time.Now()

			if client.failureCount >= 5 {
				log.Printf("Removed unresponsive client %d from channel %d", client.userId, client.channelId)
				client.close()
				ch.RemoveClient(channelId, client.conn)
			}
		}
	}
}

var channelsHub = NewChannelsHub()

func ChatHandler(service chat.Service) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		token := conn.Query("token")
		if token == "" {
			sendWSError(conn, "Token is required")
			return
		}

		userId, err := validateToken(token)
		if err != nil {
			sendWSError(conn, err.Error())
			return
		}

		// Cast channelId to int64
		channelId, err := strconv.ParseInt(conn.Params("channelId"), 10, 64)
		if err != nil {
			sendWSError(conn, "Internal error")
			return
		}

		// Check if user is a member of the channel
		exists, err := service.CheckUserMembership(int(channelId), userId)
		if err != nil || !exists {
			log.Println("error checking membership:", err)
			sendWSError(conn, "You are not a member of this channel")
			return
		}

		// Create a new client
		client := NewClient(conn, userId, int(channelId))

		// Add client to the channel
		channelsHub.AddClient(channelId, conn, client)

		log.Printf("User %d joined channel %d from IP %s\n", userId, channelId, conn.RemoteAddr().String())

		go client.writePump()

		// Remove client from the channel
		defer func() {
			log.Printf("User %d left channel %d from IP %s\n", userId, channelId, conn.RemoteAddr().String())
			client.close()

			channelsHub.RemoveClient(channelId, conn)
		}()

		for {
			msg := Message{}
			err := conn.ReadJSON(&msg.Body)
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived) {
					log.Printf("Client disconnected normally: %v", err)
					break
				}
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived,
				) {
					log.Printf("Unexpected close error: %v", err)
					break
				}
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

			// Insert message into database
			insertedMessage, err := service.InsertMessage(int(channelId), userId, msg.Body)
			if err != nil {
				log.Println("error inserting message:", err)
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			// Query user from database to populate the message
			user, err := service.FetchUserById(userId)
			if err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}
			insertedMessage.User = user

			client.send <- Result{
				Success: true,
				Message: "Message sent successfully",
			}

			channelsHub.BroadcastMessage(channelId, insertedMessage)
		}
	})
}
