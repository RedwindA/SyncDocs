package syncer

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"syncdocs/internal/database"
	gh "syncdocs/internal/github" // Alias github package
)

// Syncer handles the logic for synchronizing repository documents.
type Syncer struct {
	Store        *database.RepositoryStore
	GithubClient *gh.Client
	syncing      map[int]bool // Tracks repositories currently being synced
	mu           sync.Mutex   // Protects the syncing map
}

// NewSyncer creates a new Syncer instance.
func NewSyncer(store *database.RepositoryStore, ghClient *gh.Client) *Syncer {
	return &Syncer{
		Store:        store,
		GithubClient: ghClient,
		syncing:      make(map[int]bool),
	}
}

// SyncRepositoryByID performs the synchronization process for a single repository.
// It fetches files, aggregates content, and updates the database.
func (s *Syncer) SyncRepositoryByID(ctx context.Context, id int) error {
	s.mu.Lock()
	if s.syncing[id] {
		s.mu.Unlock()
		log.Printf("Sync already in progress for repository ID: %d. Skipping.", id)
		return fmt.Errorf("sync already in progress for repository %d", id)
	}
	s.syncing[id] = true
	s.mu.Unlock()

	// Ensure the syncing status is cleared when the function exits
	defer func() {
		s.mu.Lock()
		delete(s.syncing, id)
		s.mu.Unlock()
		log.Printf("Finished sync process for repository ID: %d", id)
	}()

	log.Printf("Starting sync for repository ID: %d", id)

	// 1. Mark as syncing in DB
	err := s.Store.UpdateSyncStatus(ctx, id, "syncing", nil)
	if err != nil {
		// Log error but proceed if possible, maybe the repo was deleted concurrently?
		log.Printf("Error marking repo %d as syncing: %v. Proceeding with sync attempt.", id, err)
		// If the error indicates the repo doesn't exist, we should stop.
		if strings.Contains(err.Error(), "not found") { // Basic check
			return fmt.Errorf("repository %d not found to start sync", id)
		}
	}

	// 2. Get repository details from DB
	repo, err := s.Store.GetRepositoryByID(ctx, id) // Get full details needed for sync
	if err != nil {
		log.Printf("Error fetching repository %d details for sync: %v", id, err)
		_ = s.Store.UpdateSyncStatus(ctx, id, "failed", fmt.Errorf("failed to fetch repository details: %w", err))
		return err
	}

	// 3. Get file list from GitHub
	log.Printf("Fetching file list for %s/%s (branch: %s) path %s", repo.Owner, repo.RepoName, repo.Branch, repo.DocsPath)
	filesInfo, err := s.GithubClient.GetRepoContentsRecursive(ctx, repo.Owner, repo.RepoName, repo.DocsPath, repo.Branch)
	if err != nil {
		log.Printf("Error getting repo contents for %d (branch: %s): %v", id, repo.Branch, err)
		_ = s.Store.UpdateSyncStatus(ctx, id, "failed", fmt.Errorf("failed to list GitHub repository contents (branch: %s): %w", repo.Branch, err))
		return err
	}
	log.Printf("Found %d potential files/dirs for repo %d (branch: %s)", len(filesInfo), id, repo.Branch)


	// 4. Filter files by extension
	allowedExtensions := make(map[string]bool)
	for _, ext := range strings.Split(repo.Extensions, ",") {
		trimmedExt := strings.TrimSpace(ext)
		if trimmedExt != "" {
			// Store with leading dot for easier matching with filepath.Ext
			allowedExtensions["."+trimmedExt] = true
		}
	}

	var filesToFetch []gh.FileInfo
	for _, fileInfo := range filesInfo {
		ext := strings.ToLower(filepath.Ext(fileInfo.Path))
		if allowedExtensions[ext] {
			filesToFetch = append(filesToFetch, fileInfo)
		}
	}
	log.Printf("Filtered down to %d files with allowed extensions for repo %d", len(filesToFetch), id)


	if len(filesToFetch) == 0 {
		log.Printf("No files with allowed extensions found for repo %d. Sync successful (empty).", id)
		err = s.Store.UpdateSyncSuccess(ctx, id, "") // Store empty content
		if err != nil {
			log.Printf("Error updating sync success (empty) for repo %d: %v", id, err)
			// Don't necessarily mark as failed, but log the update error
		}
		return nil // Successful sync, just no matching files
	}

	// 5. Sort files by path
	sort.Slice(filesToFetch, func(i, j int) bool {
		return filesToFetch[i].Path < filesToFetch[j].Path
	})

	// 6. Fetch content and aggregate
	var aggregatedContent strings.Builder
	totalFilesFetched := 0
	for _, fileInfo := range filesToFetch {
		log.Printf("Fetching content for file: %s (Repo ID: %d, Branch: %s)", fileInfo.Path, id, repo.Branch)
		// Add a timeout to individual file fetches?
		fileCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30-second timeout per file
		content, err := s.GithubClient.GetFileContent(fileCtx, repo.Owner, repo.RepoName, fileInfo.Path, repo.Branch)
		cancel() // Release context resources promptly

		if err != nil {
			log.Printf("Error getting file content for %s (Repo ID: %d, Branch: %s): %v", fileInfo.Path, id, repo.Branch, err)
			// Decide whether to fail the whole sync or just skip this file
			// For now, fail the whole sync on any file error
			_ = s.Store.UpdateSyncStatus(ctx, id, "failed", fmt.Errorf("failed to get content for file '%s' (branch: %s): %w", fileInfo.Path, repo.Branch, err))
			return err
		}

		// Add separator and content
		aggregatedContent.WriteString("---\n")
		aggregatedContent.WriteString(fmt.Sprintf("File: %s\n", fileInfo.Path))
		aggregatedContent.WriteString("---\n\n")
		aggregatedContent.WriteString(content)
		aggregatedContent.WriteString("\n\n\n") // Add extra newlines between files
		totalFilesFetched++
	}

	// 7. Update database with aggregated content
	log.Printf("Successfully fetched content for %d files for repo %d. Updating database.", totalFilesFetched, id)
	finalContent := aggregatedContent.String()
	err = s.Store.UpdateSyncSuccess(ctx, id, finalContent)
	if err != nil {
		log.Printf("Error updating sync success data for repo %d: %v", id, err)
		// Don't mark as failed if content was fetched but DB update failed, but log it.
		// The status might remain 'syncing' or revert based on previous state.
		// Consider how to handle DB update failures more robustly.
		return err // Return the DB error
	}

	log.Printf("Sync successful for repository ID: %d", id)
	return nil
}

// SyncAllRepositories iterates through all configured repositories and triggers their sync.
// This is intended to be called by a scheduler.
// Consider adding concurrency limits if syncing many repos.
func (s *Syncer) SyncAllRepositories(ctx context.Context) {
	log.Println("Starting scheduled sync for all repositories...")
	repos, err := s.Store.GetAllRepositoriesForSync(ctx)
	if err != nil {
		log.Printf("Error retrieving repositories for scheduled sync: %v", err)
		return
	}

	log.Printf("Found %d repositories to potentially sync.", len(repos))

	var wg sync.WaitGroup
	// Simple concurrency limit: process N at a time
	concurrencyLimit := 5 // Adjust as needed
	semaphore := make(chan struct{}, concurrencyLimit)


	for _, repo := range repos {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore slot

		go func(r database.Repository) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore slot

			// Create a new context for each sync operation? Or use the main one?
			// Using the main context means if it's cancelled, all ongoing syncs might stop.
			syncCtx := ctx // Use the passed context for now

			err := s.SyncRepositoryByID(syncCtx, r.ID)
			if err != nil {
				// Error is already logged within SyncRepositoryByID
				log.Printf("Scheduled sync for repo ID %d completed with error: %v", r.ID, err)
			} else {
				log.Printf("Scheduled sync for repo ID %d completed successfully.", r.ID)
			}
		}(repo)
	}

	wg.Wait() // Wait for all goroutines to finish
	log.Println("Finished scheduled sync run for all repositories.")
}
