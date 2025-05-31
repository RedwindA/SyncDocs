package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	gh "syncdocs/internal/github" // Alias github package
)

// RepositoryStore handles database operations for repositories.
type RepositoryStore struct {
	db *pgxpool.Pool
}

// NewRepositoryStore creates a new RepositoryStore.
func NewRepositoryStore(db *pgxpool.Pool) *RepositoryStore {
	return &RepositoryStore{db: db}
}

// CreateRepository inserts a new repository record into the database.
// It parses the owner and repo name from the URL.
func (s *RepositoryStore) CreateRepository(ctx context.Context, payload RepositoryCreatePayload, branchToStore string) (*Repository, error) {
	owner, repoName, err := gh.ParseRepoURL(payload.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Normalize extensions (remove spaces, ensure lowercase)
	extensions := strings.ToLower(strings.ReplaceAll(payload.Extensions, " ", ""))

	query := `
		INSERT INTO repositories (url, owner, repo_name, docs_path, extensions, branch, last_sync_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, url, owner, repo_name, docs_path, extensions, branch, aggregated_content, last_sync_status, last_sync_time, last_sync_error, created_at, updated_at
	`
	var repo Repository
	err = s.db.QueryRow(ctx, query,
		payload.URL,
		owner,
		repoName,
		payload.DocsPath,
		extensions,
		branchToStore, // Use the determined branch
		"pending",   // Initial status
	).Scan(
		&repo.ID,
		&repo.URL,
		&repo.Owner,
		&repo.RepoName,
		&repo.DocsPath,
		&repo.Extensions,
		&repo.Branch,
		&repo.AggregatedContent,
		&repo.LastSyncStatus,
		&repo.LastSyncTime,
		&repo.LastSyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return nil, fmt.Errorf("repository with URL '%s' already exists", payload.URL)
		}
		log.Printf("Error creating repository in DB: %v", err)
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &repo, nil
}

// GetRepositoryByID retrieves a single repository by its ID.
func (s *RepositoryStore) GetRepositoryByID(ctx context.Context, id int) (*Repository, error) {
	query := `
		SELECT id, url, owner, repo_name, docs_path, extensions, branch, aggregated_content, last_sync_status, last_sync_time, last_sync_error, created_at, updated_at
		FROM repositories
		WHERE id = $1
	`
	var repo Repository
	err := s.db.QueryRow(ctx, query, id).Scan(
		&repo.ID,
		&repo.URL,
		&repo.Owner,
		&repo.RepoName,
		&repo.DocsPath,
		&repo.Extensions,
		&repo.Branch,
		&repo.AggregatedContent,
		&repo.LastSyncStatus,
		&repo.LastSyncTime,
		&repo.LastSyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository with ID %d not found", id)
		}
		log.Printf("Error getting repository by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return &repo, nil
}

// ListRepositories retrieves a list of all repositories (without aggregated content).
func (s *RepositoryStore) ListRepositories(ctx context.Context) ([]RepositoryListItem, error) {
	query := `
		SELECT id, url, docs_path, extensions, branch, last_sync_status, last_sync_time, last_sync_error, updated_at
		FROM repositories
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("Error listing repositories: %v", err)
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	var items []RepositoryListItem
	for rows.Next() {
		var item RepositoryListItem
		var lastSyncError sql.NullString // Scan into NullString first
		var branch sql.NullString        // Scan branch into NullString

		err := rows.Scan(
			&item.ID,
			&item.URL,
			&item.DocsPath,
			&item.Extensions,
			&branch, // Scan into sql.NullString
			&item.LastSyncStatus,
			&item.LastSyncTime,
			&lastSyncError, // Scan into NullString
			&item.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning repository row: %v", err)
			// Decide whether to skip this row or return an error
			continue // Skip problematic row for now
		}
		item.LastSyncError = lastSyncError.String // Convert NullString to string
		item.Branch = branch.String              // Convert NullString to string for branch
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating repository rows: %v", err)
		return nil, fmt.Errorf("failed during repository list iteration: %w", err)
	}

	return items, nil
}

// UpdateRepository updates the configuration of an existing repository.
func (s *RepositoryStore) UpdateRepository(ctx context.Context, id int, payload RepositoryUpdatePayload) (*Repository, error) {
	// Normalize extensions
	extensions := strings.ToLower(strings.ReplaceAll(payload.Extensions, " ", ""))

	query := `
		UPDATE repositories
		SET docs_path = $1, extensions = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, url, owner, repo_name, docs_path, extensions, branch, aggregated_content, last_sync_status, last_sync_time, last_sync_error, created_at, updated_at
	`
	var repo Repository
	err := s.db.QueryRow(ctx, query,
		payload.DocsPath,
		extensions,
		time.Now(), // Explicitly set updated_at, though trigger should handle it
		id,
	).Scan(
		&repo.ID,
		&repo.URL,
		&repo.Owner,
		&repo.RepoName,
		&repo.DocsPath,
		&repo.Extensions,
		&repo.Branch,
		&repo.AggregatedContent,
		&repo.LastSyncStatus,
		&repo.LastSyncTime,
		&repo.LastSyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository with ID %d not found for update", id)
		}
		log.Printf("Error updating repository ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}

	return &repo, nil
}

// DeleteRepository removes a repository record from the database.
func (s *RepositoryStore) DeleteRepository(ctx context.Context, id int) error {
	query := `DELETE FROM repositories WHERE id = $1`
	commandTag, err := s.db.Exec(ctx, query, id)

	if err != nil {
		log.Printf("Error deleting repository ID %d: %v", id, err)
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("repository with ID %d not found for deletion", id)
	}

	return nil
}

// --- Methods for Syncer ---

// UpdateSyncStatus updates the sync status and error message for a repository.
func (s *RepositoryStore) UpdateSyncStatus(ctx context.Context, id int, status string, syncError error) error {
	var errMsg sql.NullString
	if syncError != nil {
		errMsg = sql.NullString{String: syncError.Error(), Valid: true}
	}

	query := `
		UPDATE repositories
		SET last_sync_status = $1, last_sync_error = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := s.db.Exec(ctx, query, status, errMsg, id)
	if err != nil {
		log.Printf("Error updating sync status for repo ID %d: %v", id, err)
		return fmt.Errorf("failed to update sync status: %w", err)
	}
	return nil
}

// UpdateSyncSuccess updates the repository content and marks the sync as successful.
func (s *RepositoryStore) UpdateSyncSuccess(ctx context.Context, id int, content string) error {
	query := `
		UPDATE repositories
		SET aggregated_content = $1, last_sync_status = $2, last_sync_time = $3, last_sync_error = NULL, updated_at = NOW()
		WHERE id = $4
	`
	_, err := s.db.Exec(ctx, query, content, "success", time.Now(), id)
	if err != nil {
		log.Printf("Error updating sync success for repo ID %d: %v", id, err)
		return fmt.Errorf("failed to update sync success data: %w", err)
	}
	return nil
}

// GetAllRepositoriesForSync retrieves all repositories to be processed by the syncer.
func (s *RepositoryStore) GetAllRepositoriesForSync(ctx context.Context) ([]Repository, error) {
	query := `
		SELECT id, url, owner, repo_name, docs_path, extensions, branch
		FROM repositories
		ORDER BY id ASC -- Or any other order suitable for processing
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("Error getting all repositories for sync: %v", err)
		return nil, fmt.Errorf("failed to get repositories for sync: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		err := rows.Scan(
			&repo.ID,
			&repo.URL,
			&repo.Owner,
			&repo.RepoName,
			&repo.DocsPath,
			&repo.Extensions,
			&repo.Branch,
		)
		if err != nil {
			log.Printf("Error scanning repository row for sync: %v", err)
			continue // Skip problematic row
		}
		repos = append(repos, repo)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating repository rows for sync: %v", err)
		return nil, fmt.Errorf("failed during repository list iteration for sync: %w", err)
	}

	return repos, nil
}
