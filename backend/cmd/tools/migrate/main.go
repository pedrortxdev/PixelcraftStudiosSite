// Tool to run database migrations
// Usage: go run cmd/tools/migrate/main.go
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/pixelcraft/api/internal/config"
	"github.com/pixelcraft/api/internal/database"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize database connection
	db, err := database.NewPostgresDB(cfg.Database, cfg.CPFEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established")

	// Run migrations
	log.Println("🚀 Running database migrations...")
	if err := database.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("Fatal error running DB migrations: %v", err)
	}

	log.Println("✅ All migrations completed successfully!")
	fmt.Println("\n📊 Migrations applied:")
	fmt.Println("   - 007_create_file_permission_tables.sql")
	fmt.Println("   - 008_update_check_file_access_function.sql")
	fmt.Println("\n✅ Database is up to date!")
}
