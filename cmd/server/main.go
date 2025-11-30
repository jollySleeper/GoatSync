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

	// 6. Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	itemRepo := repository.NewItemRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	log.Println("Repositories initialized")

	// 7. Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	collectionService := service.NewCollectionService(collectionRepo, cfg)
	itemService := service.NewItemService(itemRepo, nil, collectionRepo, memberRepo)
	memberService := service.NewMemberService(memberRepo, collectionRepo)
	invitationService := service.NewInvitationService(invitationRepo, memberRepo, userRepo)
	chunkService := service.NewChunkService(chunkRepo, collectionRepo, memberRepo, fileStorage)
	log.Println("Services initialized")

	// 8. Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	collectionHandler := handler.NewCollectionHandler(collectionService)
	itemHandler := handler.NewItemHandler(itemService)
	memberHandler := handler.NewMemberHandler(memberService)
	invitationHandler := handler.NewInvitationHandler(invitationService)
	chunkHandler := handler.NewChunkHandler(chunkService)
	websocketHandler := handler.NewWebSocketHandler()
	log.Println("Handlers initialized")

	// 9. Create and start server
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
	)

	log.Printf("Starting GoatSync server on port %s", cfg.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
