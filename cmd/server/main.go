package main

import (
	"fmt"
	"log"
	"os"

	"goatsync/internal/config"
	"goatsync/internal/database"
	"goatsync/internal/handler"
	"goatsync/internal/model"
	"goatsync/internal/repository"
	"goatsync/internal/server"
	"goatsync/internal/service"
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

	// 4. Initialize repositories
	var userRepo repository.UserRepository
	var tokenRepo repository.TokenRepository
	if db != nil {
		userRepo = repository.NewUserRepository(db)
		tokenRepo = repository.NewTokenRepository(db)
		log.Println("Using PostgreSQL repositories")
	} else {
		// Fall back to in-memory repositories for development
		log.Println("Using in-memory repositories (development mode)")
		// For now, we'll just log a warning - in-memory repos need to be implemented
		// This allows the server to start without a database for testing
		os.Exit(1) // Remove this line once in-memory repos are implemented
	}

	// 5. Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)

	// 6. Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	// 7. Create and start server
	srv := server.New(cfg, authService, authHandler)

	log.Printf("Starting GoatSync server on port %s", cfg.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
