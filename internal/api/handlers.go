package api

import (
	"context" // Import context
	// "errors" // Removed unused import
	"fmt"
	"io" // Import for io.Copy
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	// "github.com/jackc/pgx/v5" // Removed unused import

	"syncdocs/internal/database"
	"syncdocs/internal/syncer" // Import syncer
)

// API holds dependencies for API handlers.
type API struct {
	Store  *database.RepositoryStore
	Syncer *syncer.Syncer // Add Syncer dependency
}

// NewAPI creates a new API instance with dependencies.
func NewAPI(store *database.RepositoryStore, syncer *syncer.Syncer) *API {
	return &API{
		Store:  store,
		Syncer: syncer, // Assign Syncer
	}
}

// ErrorResponse represents a standard JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Repository Handlers ---

// CreateRepositoryHandler handles POST /api/repositories requests.
func (a *API) CreateRepositoryHandler(c *gin.Context) {
	var payload database.RepositoryCreatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation for extensions (comma-separated, no empty parts)
	extParts := strings.Split(payload.Extensions, ",")
	validExtensions := []string{}
	for _, ext := range extParts {
		trimmedExt := strings.TrimSpace(ext)
		if trimmedExt != "" {
			validExtensions = append(validExtensions, trimmedExt)
		}
	}
	if len(validExtensions) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Extensions cannot be empty or contain only commas/spaces"})
		return
	}
	payload.Extensions = strings.Join(validExtensions, ",") // Use cleaned extensions

	repo, err := a.Store.CreateRepository(c.Request.Context(), payload)
	if err != nil {
		// Check for specific "already exists" error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error creating repository: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create repository"})
		}
		return
	}

	// Trigger initial sync in the background
	go func(repoID int) {
		log.Printf("Triggering initial sync for newly created repository ID: %d", repoID)
		// Use context.Background for the background task
		err := a.Syncer.SyncRepositoryByID(context.Background(), repoID)
		if err != nil {
			log.Printf("Error during initial sync for repo ID %d: %v", repoID, err)
			// Note: Status might already be 'failed' if SyncRepositoryByID updated it.
			// We could potentially update status here again, but it might conflict.
		} else {
			log.Printf("Initial sync completed for repo ID %d.", repoID)
		}
	}(repo.ID) // Pass the new repo ID to the goroutine

	// Return the created repository details (consider using ListItem for consistency)
	listItem := database.RepositoryListItem{
		ID:             repo.ID,
		URL:            repo.URL,
		DocsPath:       repo.DocsPath,
		Extensions:     repo.Extensions,
		LastSyncStatus: repo.LastSyncStatus,
		LastSyncTime:   repo.LastSyncTime,
		LastSyncError:  repo.LastSyncError.String, // Convert NullString
		UpdatedAt:      repo.UpdatedAt,
	}

	c.JSON(http.StatusCreated, listItem)
}

// ListRepositoriesHandler handles GET /api/repositories requests.
func (a *API) ListRepositoriesHandler(c *gin.Context) {
	repos, err := a.Store.ListRepositories(c.Request.Context())
	if err != nil {
		log.Printf("Error listing repositories: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve repositories"})
		return
	}

	// Return empty list if no repositories found, not an error
	if repos == nil {
		repos = []database.RepositoryListItem{}
	}

	c.JSON(http.StatusOK, repos)
}

// GetRepositoryHandler handles GET /api/repositories/:id requests.
// This one returns the full details including content.
func (a *API) GetRepositoryHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid repository ID format"})
		return
	}

	repo, err := a.Store.GetRepositoryByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error getting repository %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve repository"})
		}
		return
	}

	// Return the full repository details
	c.JSON(http.StatusOK, repo)
}

// UpdateRepositoryHandler handles PUT /api/repositories/:id requests.
func (a *API) UpdateRepositoryHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid repository ID format"})
		return
	}

	var payload database.RepositoryUpdatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload: " + err.Error()})
		return
	}

	// Basic validation for extensions
	extParts := strings.Split(payload.Extensions, ",")
	validExtensions := []string{}
	for _, ext := range extParts {
		trimmedExt := strings.TrimSpace(ext)
		if trimmedExt != "" {
			validExtensions = append(validExtensions, trimmedExt)
		}
	}
	if len(validExtensions) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Extensions cannot be empty or contain only commas/spaces"})
		return
	}
	payload.Extensions = strings.Join(validExtensions, ",") // Use cleaned extensions

	repo, err := a.Store.UpdateRepository(c.Request.Context(), id, payload)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error updating repository %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update repository"})
		}
		return
	}

	// Return updated details (consider ListItem)
	listItem := database.RepositoryListItem{
		ID:             repo.ID,
		URL:            repo.URL,
		DocsPath:       repo.DocsPath,
		Extensions:     repo.Extensions,
		LastSyncStatus: repo.LastSyncStatus,
		LastSyncTime:   repo.LastSyncTime,
		LastSyncError:  repo.LastSyncError.String,
		UpdatedAt:      repo.UpdatedAt,
	}
	c.JSON(http.StatusOK, listItem)
}

// DeleteRepositoryHandler handles DELETE /api/repositories/:id requests.
func (a *API) DeleteRepositoryHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid repository ID format"})
		return
	}

	err = a.Store.DeleteRepository(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error deleting repository %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete repository"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Repository %d deleted successfully", id)})
}

// DownloadRepositoryContentHandler handles GET /api/repositories/:id/download requests.
func (a *API) DownloadRepositoryContentHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid repository ID format"})
		return
	}

	repo, err := a.Store.GetRepositoryByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error getting repository %d for download: %v", id, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve repository content"})
		}
		return
	}

	if !repo.AggregatedContent.Valid || repo.AggregatedContent.String == "" {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "No aggregated content available for this repository yet. Please sync first."})
		return
	}

	// Set headers for file download
	// Use repo name and docs path for a more descriptive filename
	filename := fmt.Sprintf("%s_%s_docs.md", repo.RepoName, strings.ReplaceAll(repo.DocsPath, "/", "_"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/markdown; charset=utf-8") // Correct MIME type for Markdown

	// Set status code
	c.Status(http.StatusOK)

	// Stream the content
	reader := strings.NewReader(repo.AggregatedContent.String)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// Log the error, as headers and status might have already been sent
		log.Printf("Error streaming repository content for ID %d: %v", id, err)
		// Attempt to write an error to the client if possible, though it might not work
		// if headers/status already sent. Gin might handle this by aborting.
		// c.Error(err) // This might be too late or inappropriate here.
	}
}

// TriggerSyncHandler handles POST /api/repositories/:id/sync requests.
func (a *API) TriggerSyncHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid repository ID format"})
		return
	}

	// Check if repo exists before triggering async sync
	_, err = a.Store.GetRepositoryByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		} else {
			log.Printf("Error finding repository %d before triggering sync: %v", id, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to find repository"})
		}
		return
	}

	// Trigger sync in a separate goroutine so the API call returns immediately.
	// Use a new context that isn't tied to the HTTP request's lifetime for the background task.
	go func() {
		// Create a background context for the sync task
		syncCtx := context.Background()
		err := a.Syncer.SyncRepositoryByID(syncCtx, id)
		if err != nil {
			// Error is logged within SyncRepositoryByID, but maybe log completion status here too.
			log.Printf("Background sync for repo ID %d finished with error: %v", id, err)
		} else {
			log.Printf("Background sync for repo ID %d finished successfully.", id)
		}
	}()

	// Return Accepted immediately
	c.JSON(http.StatusAccepted, gin.H{"message": fmt.Sprintf("Sync initiated for repository %d. Status will be updated.", id)})
}
