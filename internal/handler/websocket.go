package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

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
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		tickets: make(map[string]*Ticket),
	}
}

// CreateTicket creates a new ticket for WebSocket connection
func (h *WebSocketHandler) CreateTicket(userID, collectionID uint) string {
	h.ticketMutex.Lock()
	defer h.ticketMutex.Unlock()

	// Generate ticket ID
	ticketID := generateTicketID()

	h.tickets[ticketID] = &Ticket{
		UserID:       userID,
		CollectionID: collectionID,
		CreatedAt:    time.Now(),
	}

	return ticketID
}

// ValidateTicket validates and consumes a ticket
func (h *WebSocketHandler) ValidateTicket(ticketID string) *Ticket {
	h.ticketMutex.Lock()
	defer h.ticketMutex.Unlock()

	ticket, exists := h.tickets[ticketID]
	if !exists {
		return nil
	}

	// Check expiry (24 hours)
	if time.Since(ticket.CreatedAt) > 24*time.Hour {
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
	ticket := h.ValidateTicket(ticketID)
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
	h.handleConnection(conn, ticket)
}

func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, ticket *Ticket) {
	// Set up ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Read messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Echo back for now (TODO: implement proper message handling)
		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}
}

func generateTicketID() string {
	// Simple ticket generation (in production, use crypto/rand)
	return time.Now().Format("20060102150405.000000000")
}

