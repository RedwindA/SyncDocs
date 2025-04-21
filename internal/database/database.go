package database

import (
	"context"
	"fmt"
	"log"
	// "strings" // Removed unused import
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL script for initializing the database schema.
// Ideally, this would be read from the .sql file, but embedding for simplicity here.
const initSchemaSQL = `-- Create the repositories table
CREATE TABLE IF NOT EXISTS repositories (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL UNIQUE,       -- GitHub 仓库 URL (e.g., https://github.com/owner/repo)
    owner VARCHAR(255) NOT NULL,            -- 从 URL 解析出的 owner
    repo_name VARCHAR(255) NOT NULL,        -- 从 URL 解析出的 repo name
    docs_path VARCHAR(255) NOT NULL,        -- 文档目录路径
    extensions VARCHAR(100) NOT NULL,       -- 文件扩展名 (逗号分隔, e.g., "md,mdx")
    aggregated_content TEXT,                -- 合并后的文档内容
    last_sync_status VARCHAR(50) DEFAULT 'pending', -- 同步状态: pending, success, failed, syncing
    last_sync_time TIMESTAMPTZ,             -- 上次成功同步时间
    last_sync_error TEXT,                   -- 上次同步错误信息
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Function to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to call the function before update
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_repositories_updated_at' AND tgrelid = 'repositories'::regclass) THEN
        CREATE TRIGGER update_repositories_updated_at
        BEFORE UPDATE ON repositories
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
    END IF;
END
$$;

-- Add comments to columns for better understanding (optional, but good practice)
-- These might fail if run multiple times but are generally safe with IF NOT EXISTS or similar checks implicitly handled by COMMENT ON
-- COMMENT ON COLUMN repositories.url IS 'GitHub repository URL (e.g., https://github.com/owner/repo)';
-- COMMENT ON COLUMN repositories.owner IS 'Owner part extracted from the URL';
-- COMMENT ON COLUMN repositories.repo_name IS 'Repository name part extracted from the URL';
-- COMMENT ON COLUMN repositories.docs_path IS 'Path to the documentation directory within the repository';
-- COMMENT ON COLUMN repositories.extensions IS 'Comma-separated list of file extensions to sync (e.g., md,mdx)';
-- COMMENT ON COLUMN repositories.aggregated_content IS 'Concatenated content of all synced files';
-- COMMENT ON COLUMN repositories.last_sync_status IS 'Status of the last synchronization attempt (pending, success, failed, syncing)';
-- COMMENT ON COLUMN repositories.last_sync_time IS 'Timestamp of the last successful synchronization';
-- COMMENT ON COLUMN repositories.last_sync_error IS 'Error message from the last failed synchronization';
`

// ConnectDB establishes a connection pool to the PostgreSQL database.
func ConnectDB(databaseURL string) (*pgxpool.Pool, error) {
	log.Println("Connecting to database...")

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// You might want to configure pool settings here, e.g.,
	// config.MaxConns = 10
	// config.MinConns = 2
	// config.MaxConnLifetime = time.Hour
	// config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Close the pool if ping fails
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Database connection established successfully.")

	// Initialize schema after successful connection
	if err := InitializeSchema(ctx, pool); err != nil {
		pool.Close() // Close pool if schema initialization fails
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return pool, nil
}

// InitializeSchema creates the necessary tables and functions if they don't exist.
func InitializeSchema(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Initializing database schema...")
	// Execute the schema initialization SQL.
	// pgx Exec can handle multi-statement strings separated by semicolons.
	_, err := pool.Exec(ctx, initSchemaSQL)
	if err != nil {
		// Log the specific error but return a generic message
		log.Printf("Error executing schema initialization SQL: %v", err)
		// Check for specific errors if needed, e.g., permission denied
		return fmt.Errorf("database schema initialization failed")
	}
	log.Println("Database schema initialization successful (or already up-to-date).")
	return nil
}

// CloseDB closes the database connection pool.
// It's good practice to defer this in main.
func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		log.Println("Closing database connection pool...")
		pool.Close()
	}
}
