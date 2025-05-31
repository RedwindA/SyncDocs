package main

import (
	"context" // Keep one context
	"fmt"     // Keep one fmt
	"log"     // Keep one log
	"net/http" // Keep one net/http
	"strings" // Add missing strings import

	"github.com/gin-gonic/gin"

	"syncdocs/internal/api" // Import the api package
	"syncdocs/internal/auth"
	"syncdocs/internal/config"
	"syncdocs/internal/database"
	"syncdocs/internal/github"
	"syncdocs/internal/syncer"
	"syncdocs/internal/tasks" // Import tasks
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	dbPool, err := database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(dbPool) // Ensure DB connection is closed on exit

	// Initialize GitHub client
	// Use context.Background() for initialization, specific requests should use request context
	githubClient, err := github.NewClient(context.Background(), cfg.GithubToken)
	if err != nil {
		// Log warning instead of fatal, as unauthenticated client might still be intended
		log.Printf("Warning: Failed to initialize authenticated GitHub client: %v. Proceeding unauthenticated.", err)
		// Potentially create an unauthenticated client explicitly if needed
		// githubClient, _ = github.NewClient(context.Background(), "")
	}


	// Create RepositoryStore instance
	repoStore := database.NewRepositoryStore(dbPool)

	// Initialize Syncer
	appSyncer := syncer.NewSyncer(repoStore, githubClient)

	// Initialize and start Task Scheduler
	scheduler := tasks.NewScheduler(cfg, appSyncer)
	scheduler.Start()
	// Ensure scheduler is stopped on shutdown (though defer might not run on fatal errors)
	// A more robust solution involves signal handling for graceful shutdown.
	defer scheduler.Stop()


	router := gin.Default()

	// Setup authentication middleware
	authMiddleware := auth.BasicAuth(cfg.AuthUser, cfg.AuthPass)

	// Basic health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes group with authentication
	apiGroup := router.Group("/api", authMiddleware)
	{
		// Register API routes, passing the authenticated group and dependencies
		api.RegisterRoutes(apiGroup, repoStore, appSyncer, githubClient) // Pass repoStore, appSyncer, and githubClient
	}

	// Serve frontend static files
	// The path "./web/frontend/dist" should match the location where assets are copied in the Dockerfile
	router.Static("/assets", "./web/frontend/dist/assets") // Serve assets like CSS, JS
	router.StaticFile("/", "./web/frontend/dist/index.html") // Serve index.html for the root path

	// Handle SPA routing: For any route not matched by API or static files, serve index.html
	router.NoRoute(func(c *gin.Context) {
		// Check if the request path starts with /api/ - if so, it's a 404 for the API
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		// Otherwise, serve the index.html for frontend routing
		c.File("./web/frontend/dist/index.html")
	})


	// Start server
	port := cfg.ServerPort
	fmt.Printf("Server listening on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
