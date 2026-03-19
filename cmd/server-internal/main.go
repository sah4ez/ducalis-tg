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
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", "internal").Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis internal API")

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable")
	db, err := connectDB(dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	// Kafka connection
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	// TODO: Initialize Kafka producer
	logger.Info().Str("brokers", kafkaBrokers).Msg("Kafka configured")

	// API key for authentication
	apiKey := getEnv("INTERNAL_API_KEY", "")
	if apiKey == "" {
		logger.Warn().Msg("INTERNAL_API_KEY not set, using default")
	}

	// Initialize services
	// TODO: Implement stores and services
	// integrationStore := postgres.NewIntegrationStore(db)
	// integrationSvc := service.NewInternalService(integrationStore, logger, apiKey)
	// healthSvc := service.NewHealthService(logger)

	// Create fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Ducalis Internal API",
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
			"service": "internal",
		})
	})

	// Readiness check
	app.Get("/ready", func(c *fiber.Ctx) error {
		// TODO: Check database, redis, kafka connections
		return c.JSON(fiber.Map{
			"ready":   true,
			"checks": map[string]bool{
				"database": true,
				"redis":    true,
				"kafka":    true,
			},
		})
	})

	// API info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Ducalis Internal API",
			"version": Version,
			"endpoints": []string{
				"POST /internal/v1/sync/tasks",
				"POST /internal/v1/events",
				"GET  /internal/v1/events/subscribe",
				"GET  /internal/v1/workspaces/by-external/:source/:id",
				"GET  /internal/v1/tasks/by-external/:workspaceId/:source/:id",
				"POST /internal/v1/auth/validate",
			},
		})
	})

	// TODO: Register tg-generated handlers
	// server := transport.New(logger, transport.WithIntegrationService(integrationSvc))
	// server.RegisterHandlers(app)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	bind := getEnv("BIND", ":8083")
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
