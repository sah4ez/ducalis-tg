package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/internal/adapter"
	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/internal/storage/postgres"
	"github.com/sah4ez/ducalis-tg/internal/transport"
)

// Version is set during build
var Version = "dev"

func main() {
	// Setup logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", "internal").Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis internal API")

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable")
	db, err := postgres.New(dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	// API key for authentication
	apiKey := getEnv("INTERNAL_API_KEY", "")
	if apiKey == "" {
		logger.Warn().Msg("INTERNAL_API_KEY not set, service will accept all requests")
	}

	// Kafka connection
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	logger.Info().Str("brokers", kafkaBrokers).Msg("Kafka configured (consumer not yet wired)")

	// Initialize repositories
	integrationRepo := postgres.NewIntegrationRepository(db)

	// Initialize services
	integrationSvc := service.NewIntegrationService(integrationRepo, logger)

	// Create adapter
	integrationAdapter := adapter.NewIntegrationAdapter(integrationSvc)

	// Build tg-generated transport server
	srv := transport.New(logger,
		transport.IntegrationService(transport.NewIntegrationService(integrationAdapter)),
	)

	// Add health and readiness checks
	app := srv.Fiber()

	// API key middleware
	app.Use(func(c *fiber.Ctx) error {
		// Skip auth for health/ready endpoints
		if c.Path() == "/health" || c.Path() == "/ready" {
			return c.Next()
		}
		if apiKey != "" {
			providedKey := c.Get("X-API-Key")
			if providedKey != apiKey {
				return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
			}
		}
		return c.Next()
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
			"service": "internal",
		})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		// Check database connectivity
		dbErr := db.Ping(c.UserContext())
		return c.JSON(fiber.Map{
			"ready": dbErr == nil,
			"checks": map[string]bool{
				"database": dbErr == nil,
			},
		})
	})

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
	srv.Shutdown()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
