package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"

	redisclient "goatsync/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	Base
	redis       *redisclient.Client
	tickets     map[string]*Ticket
	ticketMutex sync.RWMutex
}

// Ticket represents a WebSocket connection ticket
type Ticket struct {
	UserID       uint
	CollectionID uint
	CreatedAt    time.Time
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(redis *redisclient.Client) *WebSocketHandler {
	return &WebSocketHandler{
		redis:   redis,
		tickets: make(map[string]*Ticket),
	}
}

// CreateTicket creates a new ticket for WebSocket connection
func (h *WebSocketHandler) CreateTicket(ctx context.Context, userID, collectionID uint) string {
	ticketID := generateTicketID()

	// Use Redis if available
	if h.redis != nil && h.redis.IsActive() {
		if err := h.redis.SetTicket(ctx, ticketID, userID); err != nil {
			log.Printf("Redis SetTicket error: %v, falling back to in-memory", err)
		} else {
			return ticketID
		}
	}

	// Fall back to in-memory
	h.ticketMutex.Lock()
	defer h.ticketMutex.Unlock()

	h.tickets[ticketID] = &Ticket{
		UserID:       userID,
		CollectionID: collectionID,
		CreatedAt:    time.Now(),
	}

	return ticketID
}

// ValidateTicket validates and consumes a ticket
func (h *WebSocketHandler) ValidateTicket(ctx context.Context, ticketID string) *Ticket {
	// Try Redis first if available
	if h.redis != nil && h.redis.IsActive() {
		userID, err := h.redis.GetAndDeleteTicket(ctx, ticketID)
		if err != nil {
			log.Printf("Redis GetAndDeleteTicket error: %v, trying in-memory", err)
		} else if userID > 0 {
			return &Ticket{UserID: userID}
		}
	}

	// Fall back to in-memory
	h.ticketMutex.Lock()
	defer h.ticketMutex.Unlock()

	ticket, exists := h.tickets[ticketID]
	if !exists {
		return nil
	}

	// Check expiry (10 seconds for tickets)
	if time.Since(ticket.CreatedAt) > 10*time.Second {
		delete(h.tickets, ticketID)
		return nil
	}

	// Consume ticket
	delete(h.tickets, ticketID)
	return ticket
}

// Handle handles WebSocket connections at GET /api/v1/ws/:ticket/
func (h *WebSocketHandler) Handle(c *gin.Context) {
	ticketID := c.Param("ticket")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ticket"})
		return
	}

	// Validate ticket
	ticket := h.ValidateTicket(c.Request.Context(), ticketID)
	if ticket == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired ticket"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Handle connection
	h.handleConnection(c.Request.Context(), conn, ticket)
}

func (h *WebSocketHandler) handleConnection(ctx context.Context, conn *websocket.Conn, ticket *Ticket) {
	// Set up ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Subscribe to Redis channel if available
	var msgChan <-chan []byte
	var cleanup func()
	if h.redis != nil && h.redis.IsActive() && ticket.CollectionID > 0 {
		channel := "col." + string(rune(ticket.CollectionID))
		msgChan, cleanup = h.redis.Subscribe(ctx, channel)
		defer cleanup()

		// Forward Redis messages to WebSocket
		go func() {
			for msg := range msgChan {
				if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
			}
		}()
	}

	// Ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Read messages
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Echo back for now
			if err := conn.WriteMessage(messageType, message); err != nil {
				return
			}
		}
	}
}

func generateTicketID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

