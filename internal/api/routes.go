package api

import (
	"github.com/gin-contrib/gzip" // Import gzip middleware
	"github.com/gin-gonic/gin"

	"syncdocs/internal/database"
	"syncdocs/internal/syncer" // Import syncer
)

// RegisterRoutes sets up the API routes for the application.
func RegisterRoutes(router *gin.RouterGroup, store *database.RepositoryStore, syncer *syncer.Syncer) { // Add syncer parameter
	// Create API instance with dependencies
	apiHandler := NewAPI(store, syncer) // Pass syncer

	// Repository routes
	repoRoutes := router.Group("/repositories")
	{
		repoRoutes.POST("", apiHandler.CreateRepositoryHandler)       // Add new repository
		repoRoutes.GET("", apiHandler.ListRepositoriesHandler)        // List all repositories
		repoRoutes.GET("/:id", apiHandler.GetRepositoryHandler)       // Get details of one repository (incl. content)
		repoRoutes.PUT("/:id", apiHandler.UpdateRepositoryHandler)    // Update repository config
		repoRoutes.DELETE("/:id", apiHandler.DeleteRepositoryHandler) // Delete repository

		// Actions for a specific repository
		repoRoutes.POST("/:id/sync", apiHandler.TriggerSyncHandler) // Manually trigger sync
		// Apply gzip compression to the download route
		repoRoutes.GET("/:id/download", gzip.Gzip(gzip.DefaultCompression), apiHandler.DownloadRepositoryContentHandler) // Download aggregated content
	}

	// Add other routes here if needed (e.g., system status)
}
