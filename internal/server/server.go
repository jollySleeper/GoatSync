// Package server provides the HTTP server setup and route registration.
package server

import (
	"log"

	"goatsync/internal/config"
	"goatsync/internal/handler"
	"goatsync/internal/middleware"
	"goatsync/internal/service"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	cfg               *config.Config
	engine            *gin.Engine
	authService       *service.AuthService
	authHandler       *handler.AuthHandler
	collectionHandler *handler.CollectionHandler
	itemHandler       *handler.ItemHandler
	memberHandler     *handler.MemberHandler
	invitationHandler *handler.InvitationHandler
	chunkHandler      *handler.ChunkHandler
	websocketHandler  *handler.WebSocketHandler
}

// New creates a new server instance
func New(
	cfg *config.Config,
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	collectionHandler *handler.CollectionHandler,
	itemHandler *handler.ItemHandler,
	memberHandler *handler.MemberHandler,
	invitationHandler *handler.InvitationHandler,
	chunkHandler *handler.ChunkHandler,
	websocketHandler *handler.WebSocketHandler,
) *Server {
	// Set Gin mode based on config
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global middleware
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(corsMiddleware(cfg.AllowedOrigins))

	return &Server{
		cfg:               cfg,
		engine:            engine,
		authService:       authService,
		authHandler:       authHandler,
		collectionHandler: collectionHandler,
		itemHandler:       itemHandler,
		memberHandler:     memberHandler,
		invitationHandler: invitationHandler,
		chunkHandler:      chunkHandler,
		websocketHandler:  websocketHandler,
	}
}

// RegisterRoutes sets up all the routes
func (s *Server) RegisterRoutes() {
	// Root is_etebase check (some clients check at root)
	s.engine.GET("/is_etebase", s.authHandler.IsEtebase)

	// API v1 group
	api := s.engine.Group("/api/v1")

	// Authentication routes (no auth required for most)
	auth := api.Group("/authentication")
	{
		auth.GET("/is_etebase/", s.authHandler.IsEtebase)
		auth.POST("/login_challenge/", s.authHandler.LoginChallenge)
		auth.POST("/login/", s.authHandler.Login)
		auth.POST("/signup/", s.authHandler.Signup)

		// These require authentication
		authRequired := auth.Group("", middleware.RequireAuth(s.authService))
		{
			authRequired.POST("/logout/", s.authHandler.Logout)
			authRequired.POST("/change_password/", s.authHandler.ChangePassword)
			authRequired.POST("/dashboard_url/", s.authHandler.DashboardURL)
		}
	}

	// Collection routes (all require auth)
	collection := api.Group("/collection", middleware.RequireAuth(s.authService))
	{
		collection.GET("/", s.collectionHandler.List)
		collection.POST("/", s.collectionHandler.Create)
		collection.GET("/:collection_uid/", s.collectionHandler.Get)

		// Item routes (nested under collection)
		collection.GET("/:collection_uid/item/", s.itemHandler.List)
		collection.GET("/:collection_uid/item/:item_uid/", s.itemHandler.Get)
		collection.POST("/:collection_uid/item/batch/", s.itemHandler.Batch)
		collection.POST("/:collection_uid/item/transaction/", s.itemHandler.Transaction)
		collection.POST("/:collection_uid/item/fetch_updates/", s.itemHandler.FetchUpdates)

		// Chunk routes
		collection.PUT("/:collection_uid/item/:item_uid/chunk/:chunk_uid/", s.chunkHandler.Upload)
		collection.GET("/:collection_uid/item/:item_uid/chunk/:chunk_uid/download/", s.chunkHandler.Download)

		// Member routes
		collection.GET("/:collection_uid/member/", s.memberHandler.List)
		collection.DELETE("/:collection_uid/member/:username/", s.memberHandler.Remove)
		collection.PATCH("/:collection_uid/member/:username/", s.memberHandler.Modify)
		collection.POST("/:collection_uid/member/leave/", s.memberHandler.Leave)
	}

	// Invitation routes (all require auth)
	invitation := api.Group("/invitation", middleware.RequireAuth(s.authService))
	{
		// Incoming invitations
		incoming := invitation.Group("/incoming")
		{
			incoming.GET("/", s.invitationHandler.ListIncoming)
			incoming.GET("/:invitation_uid/", s.invitationHandler.GetIncoming)
			incoming.DELETE("/:invitation_uid/", s.invitationHandler.RejectIncoming)
			incoming.POST("/:invitation_uid/accept/", s.invitationHandler.AcceptIncoming)
		}

		// Outgoing invitations
		outgoing := invitation.Group("/outgoing")
		{
			outgoing.GET("/", s.invitationHandler.ListOutgoing)
			outgoing.DELETE("/:invitation_uid/", s.invitationHandler.DeleteOutgoing)
			outgoing.POST("/fetch_user_profile/", s.invitationHandler.FetchUserForInvite)
		}
	}

	// WebSocket route
	ws := api.Group("/ws")
	{
		ws.GET("/:ticket/", s.websocketHandler.Handle)
	}
}

// Run starts the HTTP server
func (s *Server) Run() error {
	s.RegisterRoutes()

	addr := ":" + s.cfg.Port
	log.Printf("Server starting on %s", addr)
	return s.engine.Run(addr)
}

// corsMiddleware returns a CORS middleware
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
