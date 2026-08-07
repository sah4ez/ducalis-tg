package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/internal/storage/postgres"
)

// Version is set during build
var Version = "dev"

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", "admin").Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis admin API")

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable")
	db, err := postgres.New(dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)

	// Initialize admin service
	adminSvc := service.NewAdminService(userRepo, workspaceRepo, logger)

	// Admin API has no tg-generated transport (no AdminService contract).
	// Use fiber directly for admin endpoints.
	app := fiber.New(fiber.Config{
		AppName:               "Ducalis Admin API",
		DisableStartupMessage: true,
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
			"service": "admin",
		})
	})

	// Admin endpoints
	app.Get("/admin/v1/users", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 50)
		offset := c.QueryInt("offset", 0)
		search := c.Query("search", "")
		users, total, err := adminSvc.ListUsers(c.UserContext(), limit, offset, search)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"users": users, "total": total})
	})

	app.Get("/admin/v1/workspaces", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 50)
		offset := c.QueryInt("offset", 0)
		workspaces, total, err := adminSvc.ListWorkspaces(c.UserContext(), limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"workspaces": workspaces, "total": total})
	})

	app.Get("/admin/v1/stats", func(c *fiber.Ctx) error {
		stats, err := adminSvc.GetSystemStats(c.UserContext())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(stats)
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	bind := getEnv("BIND", ":8082")
	logger.Info().Str("bind", bind).Msg("listening")

	go func() {
		if err := app.Listen(bind); err != nil {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	logger.Info().Msg("shutting down...")
	_ = app.Shutdown()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
