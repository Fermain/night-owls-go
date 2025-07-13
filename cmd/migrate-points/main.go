package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/migration"
	"night-owls-go/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dryRun = flag.Bool("dry-run", false, "Preview changes without applying them")
	force  = flag.Bool("force", false, "Skip confirmation prompts")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if *dryRun {
		logger.Info("🔍 Running in DRY-RUN mode - no changes will be made")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	dbConn, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	queries := db.New(dbConn)
	pointsService := service.NewPointsService(queries, logger)

	migrator := migration.NewPointsMigrator(queries, pointsService, dbConn, logger, *dryRun)

	logger.Info("🎯 Starting Historical Points Migration")
	logger.Info("📊 This will award points for all historical bookings that haven't received points yet")

	// Show preview first
	summary, err := migrator.Preview(ctx)
	if err != nil {
		logger.Error("Failed to preview migration", "error", err)
		os.Exit(1)
	}

	migrator.PrintSummary(summary)

	if !*dryRun {
		if !*force {
			fmt.Print("\n❓ Do you want to proceed with awarding these points? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				logger.Info("❌ Migration cancelled by user")
				return
			}
		}

		logger.Info("🚀 Starting migration...")
		if err := migrator.Execute(ctx); err != nil {
			logger.Error("Migration failed", "error", err)
			os.Exit(1)
		}
		logger.Info("✅ Migration completed successfully!")
	}
}