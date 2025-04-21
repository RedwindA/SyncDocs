-- Create the repositories table
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

-- Optional: Add an index for faster lookups if needed
-- CREATE INDEX idx_repositories_url ON repositories(url);

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
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_repositories_updated_at') THEN
        CREATE TRIGGER update_repositories_updated_at
        BEFORE UPDATE ON repositories
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
    END IF;
END
$$;

-- Add comments to columns for better understanding
COMMENT ON COLUMN repositories.url IS 'GitHub repository URL (e.g., https://github.com/owner/repo)';
COMMENT ON COLUMN repositories.owner IS 'Owner part extracted from the URL';
COMMENT ON COLUMN repositories.repo_name IS 'Repository name part extracted from the URL';
COMMENT ON COLUMN repositories.docs_path IS 'Path to the documentation directory within the repository';
COMMENT ON COLUMN repositories.extensions IS 'Comma-separated list of file extensions to sync (e.g., md,mdx)';
COMMENT ON COLUMN repositories.aggregated_content IS 'Concatenated content of all synced files';
COMMENT ON COLUMN repositories.last_sync_status IS 'Status of the last synchronization attempt (pending, success, failed, syncing)';
COMMENT ON COLUMN repositories.last_sync_time IS 'Timestamp of the last successful synchronization';
COMMENT ON COLUMN repositories.last_sync_error IS 'Error message from the last failed synchronization';
