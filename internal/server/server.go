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
	cfg         *config.Config
	engine      *gin.Engine
	authHandler *handler.AuthHandler
	authService *service.AuthService
}

// New creates a new server instance
func New(
	cfg *config.Config,
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
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
		cfg:         cfg,
		engine:      engine,
		authHandler: authHandler,
		authService: authService,
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
	// TODO: Add collection handler
	// collection := api.Group("/collection", middleware.RequireAuth(s.authService))
	// {
	//     collection.GET("/", s.collectionHandler.List)
	//     collection.POST("/", s.collectionHandler.Create)
	//     collection.POST("/list_multi/", s.collectionHandler.ListMulti)
	//     collection.GET("/:collection_uid/", s.collectionHandler.Get)
	//     // ... more routes
	// }

	// Invitation routes (all require auth)
	// TODO: Add invitation handler

	// WebSocket route
	// TODO: Add websocket handler
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

