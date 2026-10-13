package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"caddy-manager/internal/api/routes"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/config"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
	"caddy-manager/internal/runtime"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	log.Println("[INFO] Starting Caddio...")

	// 1. Startup Configuration Validation
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Configuration validation failed: %v", err)
	}

	log.Printf("[INFO] Config loaded: AppEnv=%s, Port=%d, DBPath=%s, CaddyURL=%s",
		cfg.AppEnv, cfg.Port, cfg.DatabasePath, cfg.CaddyAdminURL)

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("[FATAL] Failed to create database directory '%s': %v", dbDir, err)
		}
	}

	// 2. Initialize Database & Migrations
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("[INFO] Database initialized with foreign keys and WAL mode.")

	// 3. Initialize Services
	authService := auth.NewService(db)
	caddyClient := caddy.NewClient(cfg.CaddyAdminURL)
	loggerService := audit.NewLogger(db)
	versionProvider := caddy.NewSystemVersionProvider(cfg.CaddyBinaryPath)
	sitesService := sites.NewService(caddyClient, loggerService)
	snapshotStore := configstore.NewSnapshotStore(db, 50)
	configService := configstore.NewService(caddyClient, versionProvider, snapshotStore, loggerService)
	runtimeService := runtime.NewService(caddyClient, versionProvider, sitesService, configService)

	// 4. Create Fiber App
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SERVER_ERROR",
					Message: err.Error(),
				},
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// 5. Setup Routes
	routes.SetupRoutes(app, db, authService, caddyClient, loggerService, versionProvider, sitesService, configService, runtimeService)

	// 6. Embedded Frontend Static Asset Handler
	setupEmbeddedFrontend(app)

	// 7. Start Server with Graceful Shutdown
	serverAddr := fmt.Sprintf(":%d", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] Caddio listening on http://0.0.0.0:%d", cfg.Port)
		if err := app.Listen(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server listener failed: %v", err)
		}
	}()

	sig := <-quit
	log.Printf("[INFO] Received signal '%v'. Shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}

	log.Println("[INFO] Caddio server stopped cleanly.")
}

func setupEmbeddedFrontend(app *fiber.App) {
	rootFS := getFrontendFS()
	if rootFS == nil {
		log.Println("[WARN] Embedded frontend assets not found. Serving API endpoints only.")
		app.Use(func(c *fiber.Ctx) error {
			path := c.Path()
			if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/ready" {
				return c.Next()
			}
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Frontend assets not embedded in this build.",
			})
		})
		return
	}

	log.Println("[INFO] Serving embedded frontend static assets from go:embed (frontend_dist)")

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()

		// Skip API, health, and ready endpoints
		if strings.HasPrefix(path, "/api/") || path == "/health" || path == "/ready" {
			return c.Next()
		}

		// Only handle GET and HEAD methods for static frontend assets
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return c.Next()
		}

		return serveEmbeddedFile(c, rootFS, path)
	})
}

func serveEmbeddedFile(c *fiber.Ctx, fsys fs.FS, reqPath string) error {
	// Clean and normalize path
	path := strings.TrimPrefix(reqPath, "/")
	if path == "" {
		path = "index.html"
	}

	// Attempt to read requested file from embedded filesystem
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		// If file doesn't exist (e.g. SPA route like /dashboard or /login), fallback to index.html
		data, err = fs.ReadFile(fsys, "index.html")
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Frontend index.html not found in embedded build")
		}
		path = "index.html"
	}

	c.Set("Content-Type", getContentType(path))
	c.Set("Cache-Control", getCacheControl(path))

	return c.Status(fiber.StatusOK).Send(data)
}

func getCacheControl(path string) string {
	if strings.HasPrefix(path, "assets/") {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache, no-store, must-revalidate"
}

func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".ico":
		return "image/x-icon"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	case ".ttf":
		return "font/ttf"
	default:
		return "application/octet-stream"
	}
}
