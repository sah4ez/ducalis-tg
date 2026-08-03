package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
)

// Version is set during build
var Version = "dev"

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", "admin").Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis admin API")

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable")
	db, err := connectDB(dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	// Initialize admin services
	// TODO: Implement stores and services
	// adminStore := postgres.NewAdminStore(db)
	// adminSvc := service.NewAdminService(adminStore, logger)
	// adminAuthSvc := service.NewAdminAuthService(adminStore, logger, getEnv("ADMIN_JWT_SECRET", "secret"))

	// Create fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Ducalis Admin API",
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		DisableStartupMessage: false,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
			"service": "admin",
		})
	})

	// API info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Ducalis Admin API",
			"version": Version,
			"endpoints": []string{
				"POST /admin/v1/auth/login",
				"GET  /admin/v1/users",
				"GET  /admin/v1/users/:id",
				"POST /admin/v1/users/:id/ban",
				"GET  /admin/v1/workspaces",
				"GET  /admin/v1/stats",
				"GET  /admin/v1/audit-log",
			},
		})
	})

	// TODO: Register tg-generated handlers
	// server := transport.New(logger, transport.WithAdminService(adminSvc))
	// server.RegisterHandlers(app)

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
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func connectDB(url string) (*sql.DB, error) {
	// TODO: Implement database connection
	return nil, nil
}
