package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Implement proper origin checking
	},
}

type WSConnection struct {
	*websocket.Conn
	CollectionUID string
	Username      string
}

func HandleWebSocket(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ticket"})
		return
	}

	// TODO: Validate ticket from Redis
	// username, err := redisClient.Get(ctx, ticket).Result()
	// if err != nil {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket"})
	//     return
	// }

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade connection"})
		return
	}

	wsConn := &WSConnection{
		Conn:          conn,
		CollectionUID: c.Param("collection_uid"),
		Username:      "username", // TODO: Get from ticket validation
	}

	go handleWebSocketConnection(wsConn)
}

func handleWebSocketConnection(conn *WSConnection) {
	defer func() { _ = conn.Close() }()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Echo the message back for now
		// TODO: Implement proper message handling
		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}
}
