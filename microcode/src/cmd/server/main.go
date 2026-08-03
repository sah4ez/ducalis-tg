package main

import (
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
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis")

	// Create fiber apps for each API
	publicApp := createFiberApp("Ducalis Public API")
	adminApp := createFiberApp("Ducalis Admin API")
	internalApp := createFiberApp("Ducalis Internal API")

	// Setup routes
	setupRoutes(publicApp, "Public API")
	setupRoutes(adminApp, "Admin API")
	setupRoutes(internalApp, "Internal API")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start servers in goroutines
	go startServer("Public API", ":8080", publicApp, logger)
	go startServer("Admin API", ":8082", adminApp, logger)
	go startServer("Internal API", ":8083", internalApp, logger)

	// Wait for shutdown signal
	<-quit
	logger.Info().Msg("shutting down servers...")
}

func createFiberApp(name string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               name,
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

	return app
}

func setupRoutes(app *fiber.App, name string) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
			"service": name,
		})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": name,
			"version": Version,
			"message": "Ducalis - Task Prioritization Platform",
		})
	})

	// TODO: Register tg-generated handlers
	// After running `make generate`, import and register:
	// transport.RegisterWorkspaceService(app, workspaceSvc)
	// transport.RegisterTaskService(app, taskSvc)
	// transport.RegisterAuthService(app, authSvc)
}

func startServer(name, bind string, app *fiber.App, logger zerolog.Logger) {
	logger.Info().Str("name", name).Str("bind", bind).Msg("starting server")
	if err := app.Listen(bind); err != nil {
		logger.Fatal().Err(err).Str("name", name).Msg("server error")
	}
}
