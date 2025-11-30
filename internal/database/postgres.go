// Package database provides PostgreSQL database connection management using GORM.
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection instance
var DB *gorm.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// Connect establishes a connection to the PostgreSQL database.
// It can accept either a connection string (DSN) or use environment variables.
func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		// Build DSN from environment variables
		dsn = BuildDSN(Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "goatsync"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),
		})
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  getLogLevel(),
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormLogger,
		SkipDefaultTransaction:                   true, // Performance optimization
		PrepareStmt:                              true, // Cache prepared statements
		DisableForeignKeyConstraintWhenMigrating: true, // Handle circular dependencies
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Store in global variable
	DB = db

	log.Println("Database connection established successfully")
	return db, nil
}

// BuildDSN constructs a PostgreSQL connection string from Config
func BuildDSN(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, cfg.TimeZone,
	)
}

// AutoMigrate runs database migrations for all models.
// Call this after establishing the database connection.
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}
	log.Println("Database migration completed successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getLogLevel returns GORM log level based on DEBUG environment variable
func getLogLevel() logger.LogLevel {
	if os.Getenv("DEBUG") == "true" {
		return logger.Info
	}
	return logger.Warn
}

// Transaction wraps a function in a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, it is committed.
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}

// Ping checks if the database connection is alive
func Ping() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

