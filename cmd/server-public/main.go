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
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", "public").Logger()
	logger.Info().Str("version", Version).Msg("starting ducalis public API")

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://ducalis:ducalis123@localhost:5432/ducalis?sslmode=disable")
	db, err := postgres.New(dbURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	memberRepo := postgres.NewMemberRepository(db)
	taskRepo := postgres.NewTaskRepository(db)
	voteRepo := postgres.NewVoteRepository(db)
	estimationRepo := postgres.NewEstimationRepository(db)

	// Initialize services
	authSvc := service.NewAuthService(userRepo, logger, jwtSecret)
	workspaceSvc := service.NewWorkspaceService(workspaceRepo, memberRepo, userRepo, logger)
	taskSvc := service.NewTaskService(taskRepo, workspaceRepo, voteRepo, estimationRepo, logger)

	// Create adapters (bridge between contract interfaces and service implementations)
	authAdapter := adapter.NewAuthAdapter(authSvc)
	workspaceAdapter := adapter.NewWorkspaceAdapter(workspaceSvc)
	taskAdapter := adapter.NewTaskAdapter(taskSvc)

	// Build tg-generated transport server
	srv := transport.New(logger,
		transport.AuthService(transport.NewAuthService(authAdapter)),
		transport.WorkspaceService(transport.NewWorkspaceService(workspaceAdapter)),
		transport.TaskService(transport.NewTaskService(taskAdapter)),
	)

	// Add health check
	app := srv.Fiber()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": Version,
			"service": "public",
		})
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	bind := getEnv("BIND", ":8080")
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
