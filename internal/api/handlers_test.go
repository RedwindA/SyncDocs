package api

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gz "github.com/gin-contrib/gzip" // Alias to avoid conflict with standard gzip
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"syncdocs/internal/database"
	// "syncdocs/internal/syncer" // Syncer not directly tested here but needed for API struct
)

// MockRepositoryStore implements the parts of RepositoryStoreInterface needed for these tests.
type MockRepositoryStore struct {
	GetRepositoryByIDFunc func(ctx *gin.Context, id int) (*database.Repository, error)
	// Add other methods if needed for other handler tests
}

func (m *MockRepositoryStore) GetRepositoryByID(ctx *gin.Context, id int) (*database.Repository, error) {
	if m.GetRepositoryByIDFunc != nil {
		return m.GetRepositoryByIDFunc(ctx, id)
	}
	return nil, errors.New("GetRepositoryByIDFunc not implemented")
}

// Unused methods for this test, but part of a broader interface potentially
func (m *MockRepositoryStore) CreateRepository(ctx *gin.Context, payload database.RepositoryCreatePayload) (*database.Repository, error) {
	return nil, nil
}
func (m *MockRepositoryStore) ListRepositories(ctx *gin.Context) ([]database.RepositoryListItem, error) {
	return nil, nil
}
func (m *MockRepositoryStore) UpdateRepository(ctx *gin.Context, id int, payload database.RepositoryUpdatePayload) (*database.Repository, error) {
	return nil, nil
}
func (m *MockRepositoryStore) DeleteRepository(ctx *gin.Context, id int) error {
	return nil
}
func (m *MockRepositoryStore) UpdateRepositorySyncStatus(ctx *gin.Context, id int, status string, syncError sql.NullString) error {
	return nil
}
func (m *MockRepositoryStore) GetRepositoryContent(ctx *gin.Context, id int) (*database.RepositoryContent, error) {
	return nil, nil
}
func (m *MockRepositoryStore) UpdateRepositoryContent(ctx *gin.Context, id int, content string, fileDetails map[string]database.FileDetail) error {
	return nil
}


func setupTestRouter(mockStore *MockRepositoryStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New() // Use gin.New() instead of gin.Default() to avoid default middleware like logger in tests unless desired
	
	// The API handler requires a Syncer, but for this specific test, it's not used.
	// We can pass nil or a very basic mock if the NewAPI function requires it.
	// For now, assuming NewAPI can handle a nil syncer or we create a dummy one.
	apiHandler := NewAPI(mockStore, nil) // Pass mockStore, Syncer can be nil if not used by this handler path

	// Setup the specific route with gzip middleware
	apiGroup := router.Group("/api")
	repoRoutes := apiGroup.Group("/repositories")
	{
		repoRoutes.GET("/:id/download", gz.Gzip(gz.DefaultCompression), apiHandler.DownloadRepositoryContentHandler)
	}
	return router
}

func TestDownloadRepositoryContent_Success(t *testing.T) {
	mockStore := &MockRepositoryStore{}
	router := setupTestRouter(mockStore)

	repoID := 1
	expectedContent := "This is some **markdown** content for testing.\nAnd another line."
	expectedRepoName := "test-repo"
	expectedDocsPath := "docs/guide"
	expectedFilename := fmt.Sprintf("%s_%s_docs.md", expectedRepoName, strings.ReplaceAll(expectedDocsPath, "/", "_"))

	mockStore.GetRepositoryByIDFunc = func(ctx *gin.Context, id int) (*database.Repository, error) {
		if id == repoID {
			return &database.Repository{
				ID:                repoID,
				RepoName:          expectedRepoName,
				DocsPath:          expectedDocsPath,
				AggregatedContent: sql.NullString{String: expectedContent, Valid: true},
				LastSyncTime:      sql.NullTime{Time: time.Now(), Valid: true},
				// Other fields can be zero/default values
			}, nil
		}
		return nil, errors.New("repository not found")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/repositories/%d/download", repoID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Equal(t, "attachment; filename="+expectedFilename, w.Header().Get("Content-Disposition"))
	assert.Equal(t, "text/markdown; charset=utf-8", w.Header().Get("Content-Type"))

	// Decompress response
	gzipReader, err := gzip.NewReader(w.Body)
	assert.NoError(t, err)
	defer gzipReader.Close()

	decompressedBody, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(decompressedBody))
}

func TestDownloadRepositoryContent_RepoNotFound(t *testing.T) {
	mockStore := &MockRepositoryStore{}
	router := setupTestRouter(mockStore)

	repoID := 99 // Non-existent ID
	mockStore.GetRepositoryByIDFunc = func(ctx *gin.Context, id int) (*database.Repository, error) {
		return nil, errors.New("database error: repository not found") // Simulate DB error
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/repositories/%d/download", repoID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var jsonResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Contains(t, jsonResponse.Error, "repository not found")
}

func TestDownloadRepositoryContent_NoContent(t *testing.T) {
	mockStore := &MockRepositoryStore{}
	router := setupTestRouter(mockStore)

	repoID := 2
	expectedRepoName := "empty-repo"
	expectedDocsPath := "docs"

	// Test case 1: AggregatedContent.Valid is false
	t.Run("NoContent_InvalidAggregatedContent", func(t *testing.T) {
		mockStore.GetRepositoryByIDFunc = func(ctx *gin.Context, id int) (*database.Repository, error) {
			if id == repoID {
				return &database.Repository{
					ID:                repoID,
					RepoName:          expectedRepoName,
					DocsPath:          expectedDocsPath,
					AggregatedContent: sql.NullString{String: "", Valid: false}, // Content is not valid
				}, nil
			}
			return nil, errors.New("repository not found")
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/repositories/%d/download", repoID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var jsonResponse ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
		assert.NoError(t, err)
		assert.Equal(t, "No aggregated content available for this repository yet. Please sync first.", jsonResponse.Error)
	})

	// Test case 2: AggregatedContent.Valid is true, but String is empty
	t.Run("NoContent_EmptyAggregatedContentString", func(t *testing.T) {
		mockStore.GetRepositoryByIDFunc = func(ctx *gin.Context, id int) (*database.Repository, error) {
			if id == repoID {
				return &database.Repository{
					ID:                repoID,
					RepoName:          expectedRepoName,
					DocsPath:          expectedDocsPath,
					AggregatedContent: sql.NullString{String: "", Valid: true}, // Content is valid but empty
				}, nil
			}
			return nil, errors.New("repository not found")
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/repositories/%d/download", repoID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var jsonResponse ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
		assert.NoError(t, err)
		assert.Equal(t, "No aggregated content available for this repository yet. Please sync first.", jsonResponse.Error)
	})
}

// TestDownloadRepositoryContent_InvalidID tests the handler's response to an invalid repository ID format.
func TestDownloadRepositoryContent_InvalidID(t *testing.T) {
	// No mock store needed as this should be caught by Gin's parameter binding/strconv.Atoi
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiHandler := NewAPI(nil, nil) // Store and Syncer not strictly needed for this path

	apiGroup := router.Group("/api")
	repoRoutes := apiGroup.Group("/repositories")
	{
		repoRoutes.GET("/:id/download", gz.Gzip(gz.DefaultCompression), apiHandler.DownloadRepositoryContentHandler)
	}
	
	w := httptest.NewRecorder()
	// Pass "invalid" as ID, which cannot be converted to an integer
	req, _ := http.NewRequest("GET", "/api/repositories/invalid/download", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var jsonResponse ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid repository ID format", jsonResponse.Error)
}

// Small helper to ensure the mock store conforms to the interface parts we care about.
// This won't be run as a test but will fail compilation if MockRepositoryStore is wrong.
func _assertMockRepositoryStoreInterface(m *MockRepositoryStore) {
	var _ interface {
		GetRepositoryByID(ctx *gin.Context, id int) (*database.Repository, error)
		// Add other methods from the actual interface if needed for compilation check
	} = m
}
