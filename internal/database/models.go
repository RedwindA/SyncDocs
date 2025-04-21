package database

import (
	"database/sql"
	"time"
)

// Repository represents a monitored GitHub repository configuration and its state.
// Corresponds to the 'repositories' table in the database.
type Repository struct {
	ID                int            `db:"id"`
	URL               string         `db:"url"`
	Owner             string         `db:"owner"`
	RepoName          string         `db:"repo_name"`
	DocsPath          string         `db:"docs_path"`
	Extensions        string         `db:"extensions"` // Comma-separated list
	AggregatedContent sql.NullString `db:"aggregated_content"` // Use sql.NullString for potentially NULL TEXT field
	LastSyncStatus    string         `db:"last_sync_status"`   // e.g., pending, success, failed, syncing
	LastSyncTime      sql.NullTime   `db:"last_sync_time"`     // Use sql.NullTime for potentially NULL TIMESTAMPTZ
	LastSyncError     sql.NullString `db:"last_sync_error"`    // Use sql.NullString for potentially NULL TEXT field
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

// RepositoryListItem represents a repository item for listing purposes,
// omitting the potentially large aggregated content.
type RepositoryListItem struct {
	ID             int          `json:"id"`
	URL            string       `json:"url"`
	DocsPath       string       `json:"docs_path"`
	Extensions     string       `json:"extensions"`
	LastSyncStatus string       `json:"last_sync_status"`
	LastSyncTime   sql.NullTime `json:"last_sync_time"` // Keep as sql.NullTime for JSON marshalling
	LastSyncError  string       `json:"last_sync_error"` // Convert NullString to string for simpler JSON
	UpdatedAt      time.Time    `json:"updated_at"`
}

// RepositoryCreatePayload defines the structure for creating a new repository entry.
type RepositoryCreatePayload struct {
	URL        string `json:"url" binding:"required,url"`
	DocsPath   string `json:"docs_path" binding:"required"`
	Extensions string `json:"extensions" binding:"required"` // e.g., "md,mdx"
}

// RepositoryUpdatePayload defines the structure for updating an existing repository entry.
type RepositoryUpdatePayload struct {
	DocsPath   string `json:"docs_path" binding:"required"`
	Extensions string `json:"extensions" binding:"required"`
}
