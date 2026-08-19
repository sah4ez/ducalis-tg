package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"github.com/sah4ez/ducalis-tg/internal/adapter"
	"github.com/sah4ez/ducalis-tg/internal/service"
	"github.com/sah4ez/ducalis-tg/internal/storage/postgres"
	"github.com/sah4ez/ducalis-tg/internal/transport"
)

// Version is set during build
var Version = "dev"

// webDist is the React build output directory (relative to CWD).
var webDist = getEnv("WEB_DIST", "./web/dist")

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

	// Build tg-generated transport server (tg v3: options take the contract
	// implementations directly; the server logs via *slog.Logger).
	// Options apply IN ORDER, so transport.Service (JWT middleware) must come
	// FIRST — fiber runs middleware/handlers in registration order, and the
	// per-service options register their routes right here inside New.
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	srv := transport.New(slogLogger,
		transport.Service(authRoutes{jwtSecret: jwtSecret}),
		transport.AuthService(authAdapter),
		transport.WorkspaceService(workspaceAdapter),
		transport.TaskService(taskAdapter),
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

	// Serve the React SPA (web/dist) — registered AFTER the API routes so
	// /api/* and /health are matched first. SPA fallback returns index.html
	// for any non-API GET so client-side routing works on refresh.
	if _, err := os.Stat(webDist); err == nil {
		app.Static("/", webDist)
		app.Use(func(c *fiber.Ctx) error {
			if c.Method() == fiber.MethodGet && !strings.HasPrefix(c.Path(), "/api/") {
				return c.SendFile(webDist + "/index.html")
			}
			return c.Next()
		})
		logger.Info().Str("dir", webDist).Msg("serving web UI")
	} else {
		logger.Warn().Str("dir", webDist).Msg("web/dist not found, API-only mode")
	}

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

// authRoutes implements transport.ServiceRoute: its SetRoutes runs inside
// transport.New (via transport.Service option) BEFORE the per-service route
// registrations, so the JWT middleware wraps every API route. The generated
// handlers pass ftx.UserContext() down to the services, so enriching it via
// c.SetUserContext delivers service.UserIDKey without touching generated code.
type authRoutes struct {
	jwtSecret string
}

func (a authRoutes) SetRoutes(app *fiber.App) {
	app.Use(a.middleware())
}

func (a authRoutes) middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		p := c.Path()
		// Guard ONLY API paths; the SPA, /health and static assets pass through.
		if !strings.HasPrefix(p, "/api/") {
			return c.Next()
		}
		// Auth endpoints are public.
		if strings.HasPrefix(p, "/api/v1/auth/register") ||
			strings.HasPrefix(p, "/api/v1/auth/login") ||
			strings.HasPrefix(p, "/api/v1/auth/oauth") {
			return c.Next()
		}
		header := c.Get("Authorization")
		if header == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}
		userID, err := parseUserID(strings.TrimPrefix(header, "Bearer "), a.jwtSecret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		c.SetUserContext(context.WithValue(c.UserContext(), service.UserIDKey, userID))
		return c.Next()
	}
}

// parseUserID validates an HS256 JWT minted by service.AuthService and returns
// its user_id claim.
func parseUserID(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	uid, _ := claims["user_id"].(string)
	if uid == "" {
		return "", fmt.Errorf("missing user_id claim")
	}
	return uid, nil
}
