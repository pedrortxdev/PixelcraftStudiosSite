package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// WSClient represents a connected client
type WSClient struct {
	Hub      *WSHub
	Conn     *websocket.Conn
	Send     chan []byte
	TicketID string // Optional: if client is subscribed to a specific ticket
	UserID   string
}

// WSHub maintains the set of active clients and broadcasts messages
type WSHub struct {
	// Registered clients
	clients map[*WSClient]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *WSClient

	// Unregister requests from clients
	unregister chan *WSClient

	// Rooms (Ticket ID -> Set of Clients)
	rooms map[string]map[*WSClient]bool
	
	mu sync.RWMutex
}

func NewWSHub() *WSHub {
	return &WSHub{
		broadcast:  make(chan []byte),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		clients:    make(map[*WSClient]bool),
		rooms:      make(map[string]map[*WSClient]bool),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.TicketID != "" {
				if h.rooms[client.TicketID] == nil {
					h.rooms[client.TicketID] = make(map[*WSClient]bool)
				}
				h.rooms[client.TicketID][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.TicketID != "" && h.rooms[client.TicketID] != nil {
					delete(h.rooms[client.TicketID], client)
					if len(h.rooms[client.TicketID]) == 0 {
						delete(h.rooms, client.TicketID)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastToRoom sends a message to all clients in a specific room (ticket)
func (h *WSHub) BroadcastToRoom(roomID string, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling WS message:", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- bytes:
			default:
				// Client disconnected
				// Cleanup is handled in Run loop usually, but here we just skip
			}
		}
	}
}

// WSHandler handles WebSocket requests
type WSHandler struct {
	Hub            *WSHub
	JWTSecret      string
	AllowedOrigins []string
}

func NewWSHandler(hub *WSHub, jwtSecret string, allowedOrigins []string) *WSHandler {
	return &WSHandler{
		Hub:            hub,
		JWTSecret:      jwtSecret,
		AllowedOrigins: allowedOrigins,
	}
}

func (h *WSHandler) ServeWS(c *gin.Context) {
	ticketID := c.Query("ticket_id")
	tokenString := c.Query("token")

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing WebSocket token"})
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid WebSocket token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id in token"})
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // allow non-browser clients mapping
			}
			
			for _, allowed := range h.AllowedOrigins {
				if allowed == "*" || origin == allowed || strings.HasSuffix(origin, allowed) {
					return true
				}
			}
			return false
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS Upgrade error:", err)
		return
	}

	client := &WSClient{
		Hub:      h.Hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		TicketID: ticketID,
		UserID:   userID,
	}

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(4096) // Max 4KB per message to prevent OOM
	
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
