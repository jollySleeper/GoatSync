package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goatsync/internal/config"
	"goatsync/internal/database"
	"goatsync/internal/handler"
	"goatsync/internal/model"
	redisclient "goatsync/internal/redis"
	"goatsync/internal/repository"
	"goatsync/internal/server"
	"goatsync/internal/service"
	"goatsync/internal/storage"
)

const banner = `
--------------------------- Welcome To ------------------------------

 ██████╗  ██████╗  █████╗ ████████╗███████╗██╗   ██╗███╗   ██╗ ██████╗
██╔════╝ ██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗ ██╔╝████╗  ██║██╔════╝
██║  ███╗██║   ██║███████║   ██║   ███████╗ ╚████╔╝ ██╔██╗ ██║██║     
██║   ██║██║   ██║██╔══██║   ██║   ╚════██║  ╚██╔╝  ██║╚██╗██║██║     
╚██████╔╝╚██████╔╝██║  ██║   ██║   ███████║   ██║   ██║ ╚████║╚██████╗
 ╚═════╝  ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝   ╚═╝   ╚═╝  ╚═══╝ ╚═════╝

----------------------------------------------------------------------
`

func main() {
	fmt.Print(banner)

	// 1. Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded (Debug: %v, Port: %s)", cfg.Debug, cfg.Port)

	// 2. Check required configuration
	if cfg.EncryptionSecret == "" {
		log.Println("WARNING: ENCRYPTION_SECRET not set. Using default (insecure for production!)")
		cfg.EncryptionSecret = "default-development-secret-key-32"
	}

	// 3. Initialize database connection
	var db = database.DB
	if cfg.DatabaseURL != "" {
		var err error
		db, err = database.Connect(cfg.DatabaseURL)
		if err != nil {
			log.Printf("WARNING: Failed to connect to database: %v", err)
			log.Println("Running in memory-only mode (data will not persist)")
		} else {
			// Run auto-migrations
			if err := database.AutoMigrate(db,
				&model.Stoken{},
				&model.User{},
				&model.UserInfo{},
				&model.AuthToken{},
				&model.CollectionType{},
				&model.Collection{},
				&model.CollectionItem{},
				&model.CollectionItemRevision{},
				&model.CollectionItemChunk{},
				&model.RevisionChunkRelation{},
				&model.CollectionMember{},
				&model.CollectionMemberRemoved{},
				&model.CollectionInvitation{},
			); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
		}
	} else {
		log.Println("WARNING: DATABASE_URL not set. Running in memory-only mode")
	}

	// 4. Check if we have a database
	if db == nil {
		log.Println("ERROR: Database connection required. Set DATABASE_URL environment variable.")
		os.Exit(1)
	}

	// 5. Initialize file storage
	fileStorage := storage.NewFileStorage(cfg.ChunkStoragePath)

	// 6. Initialize Redis (optional)
	var redis *redisclient.Client
	if cfg.RedisURL != "" {
		var err error
		redis, err = redisclient.New(cfg.RedisURL)
		if err != nil {
			log.Printf("WARNING: Failed to connect to Redis: %v", err)
			log.Println("WebSocket pub/sub will use in-memory fallback")
		} else {
			log.Println("Redis connected successfully")
		}
	}

	// 7. Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	itemRepo := repository.NewItemRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	log.Println("Repositories initialized")

	// 8. Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	collectionService := service.NewCollectionService(collectionRepo, cfg)
	itemService := service.NewItemService(itemRepo, nil, collectionRepo, memberRepo)
	memberService := service.NewMemberService(memberRepo, collectionRepo)
	invitationService := service.NewInvitationService(invitationRepo, memberRepo, userRepo)
	chunkService := service.NewChunkService(chunkRepo, collectionRepo, memberRepo, fileStorage)
	log.Println("Services initialized")

	// 9. Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	collectionHandler := handler.NewCollectionHandler(collectionService)
	itemHandler := handler.NewItemHandler(itemService)
	memberHandler := handler.NewMemberHandler(memberService)
	invitationHandler := handler.NewInvitationHandler(invitationService)
	chunkHandler := handler.NewChunkHandler(chunkService)
	websocketHandler := handler.NewWebSocketHandler(redis)
	healthHandler := handler.NewHealthHandler(db)
	log.Println("Handlers initialized")

	// 10. Create and start server
	srv := server.New(
		cfg,
		authService,
		authHandler,
		collectionHandler,
		itemHandler,
		memberHandler,
		invitationHandler,
		chunkHandler,
		websocketHandler,
		healthHandler,
	)

	// 11. Setup graceful shutdown
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: srv.Engine(),
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting GoatSync server on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Redis connection
	if redis != nil {
		if err := redis.Close(); err != nil {
			log.Printf("Redis close error: %v", err)
		}
	}

	// Close database connection
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	log.Println("Server exited gracefully")
}
