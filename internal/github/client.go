package github

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log"
	"net/http" // Import net/http package
	"net/url"
	"strings"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// Client wraps the go-github client.
type Client struct {
	*github.Client
}

// NewClient creates a new GitHub API client authenticated with a Personal Access Token (PAT).
func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		log.Println("Warning: No GITHUB_TOKEN provided. GitHub API interactions will be unauthenticated and rate-limited.")
		// Return a client without authentication
		return &Client{github.NewClient(nil)}, nil
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Optional: Verify authentication by getting the current user
	// user, _, err := client.Users.Get(ctx, "")
	// if err != nil {
	// 	// Don't fail initialization, but log a warning.
	// 	// The token might be valid but lack permissions for users:read.
	// 	log.Printf("Warning: Could not verify GitHub token. Check token validity and permissions. Error: %v", err)
	// } else {
	// 	log.Printf("GitHub client authenticated as: %s", *user.Login)
	// }

	log.Println("GitHub client initialized successfully.")
	return &Client{client}, nil
}

// ParseRepoURL extracts the owner and repository name from a GitHub URL.
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Hostname() != "github.com" {
		return "", "", fmt.Errorf("URL is not a github.com URL")
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("URL path does not contain owner and repo: %s", parsedURL.Path)
	}

	owner = pathParts[0]
	// Remove potential .git suffix
	repo = strings.TrimSuffix(pathParts[1], ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not extract owner or repo from URL path")
	}

	return owner, repo, nil
}

// FileInfo holds path and SHA for a file in the repository.
type FileInfo struct {
	Path string
	SHA  string
}

// GetRepoContentsRecursive fetches all file entries recursively starting from a given path.
// It returns a flat list of FileInfo for files only.
func (c *Client) GetRepoContentsRecursive(ctx context.Context, owner, repo, path string) ([]FileInfo, error) {
	var allFiles []FileInfo

	queue := []string{path} // Use a queue for breadth-first traversal, though depth-first works too

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]

		// Get contents of the current directory path
		// Note: Using default branch if no ref is specified. Consider adding ref/branch parameter if needed.
		_, dirContents, _, err := c.Repositories.GetContents(ctx, owner, repo, currentPath, nil)
		if err != nil {
			// Handle common errors like 404 Not Found gracefully
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("Warning: Path not found in repo %s/%s: %s", owner, repo, currentPath)
				continue // Skip this path if not found
			}
			log.Printf("Error getting contents for %s/%s path %s: %v", owner, repo, currentPath, err)
			return nil, fmt.Errorf("failed to get contents for path '%s': %w", currentPath, err)
		}

		if dirContents == nil {
			// This might happen if the path points directly to a file initially, handle it
			fileContent, _, _, err := c.Repositories.GetContents(ctx, owner, repo, currentPath, nil)
			if err == nil && fileContent != nil && fileContent.GetType() == "file" {
				allFiles = append(allFiles, FileInfo{Path: *fileContent.Path, SHA: *fileContent.SHA})
				continue // Processed the single file path
			}
			// If it's not a file or error occurred, log and continue
			log.Printf("Warning: Received nil directory contents for path %s, skipping.", currentPath)
			continue
		}


		for _, item := range dirContents {
			itemPath := *item.Path // Dereference pointer safely
			itemType := *item.Type // Dereference pointer safely

			if itemType == "dir" {
				queue = append(queue, itemPath) // Add subdirectory to the queue
			} else if itemType == "file" {
				if item.SHA == nil || item.Path == nil {
					log.Printf("Warning: Skipping file with missing SHA or Path in %s", currentPath)
					continue
				}
				allFiles = append(allFiles, FileInfo{Path: itemPath, SHA: *item.SHA})
			}
			// Ignore other types like "symlink", "submodule" for now
		}
	}

	return allFiles, nil
}


// GetFileContent fetches the raw content of a specific file using its path.
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	// GetContents can retrieve file content directly for smaller files.
	// For larger files, GetBlob might be necessary, but GetContents is simpler.
	fileContent, _, _, err := c.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		log.Printf("Error getting file content for %s/%s path %s: %v", owner, repo, path, err)
		// Handle 404 specifically
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("failed to get file content for '%s': %w", path, err)
	}

	if fileContent == nil || fileContent.GetType() != "file" {
		return "", fmt.Errorf("path is not a file: %s", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		log.Printf("Error decoding file content for %s: %v", path, err)
		return "", fmt.Errorf("failed to decode file content for '%s': %w", path, err)
	}

	return content, nil
}
