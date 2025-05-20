# SyncDocs

## Project Description

SyncDocs is a service that synchronizes documentation or files from GitHub repositories, providing a web interface to manage and view synced content. It utilizes the GitHub API for accessing repository data, stores information in a PostgreSQL database, and performs background synchronization tasks.

## Prerequisites for Deployment

*   Git
*   Docker
*   Docker Compose

## Deployment Steps

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/RedwindA/SyncDocs
    ```

2.  **Navigate to the project directory:**
    ```bash
    cd SyncDocs
    ```
    (Or the name you cloned the repository as)

3.  **Create a `.env` file:**
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```

4.  **Configure the `.env` file:**
    Open the `.env` file and update the following variables with your specific settings:
    *   `SERVER_PORT`: The port on which the application server will listen (default: `8080`).
    *   `AUTH_USER`: Username for basic authentication (e.g., `admin`). **Replace with a strong, unique username.**
    *   `AUTH_PASS`: Password for basic authentication (e.g., `changeme`). **Replace with a strong, unique password.**
    *   `DATABASE_URL`: Connection string for your PostgreSQL database.
        *   Example: `postgres://user:password@host:port/dbname?sslmode=disable`
    *   `GITHUB_TOKEN`: Your GitHub Personal Access Token. This token needs the `repo` scope to access repository contents. You can generate one at [https://github.com/settings/tokens](https://github.com/settings/tokens).
    *   `SYNC_INTERVAL`: The interval for background synchronization tasks (e.g., `1h` for 1 hour, `30m` for 30 minutes). Defaults to `1h` if not set or invalid.

5.  **Build and run the application:**
    Use Docker Compose to pull the images and start the containers in detached mode:
    ```bash
    docker compose up -d
    ```

6.  **Access the application:**
    Once the containers are running, the application will be accessible in your web browser at:
    `http://<your_host_ip_or_localhost>:${SERVER_PORT}`
    (Replace `<your_host_ip_or_localhost>` with your server's IP address or `localhost` if running locally, and `${SERVER_PORT}` with the port you configured in the `.env` file).

## Project Structure (Overview)

*   `cmd/`: Contains the main application entry points (e.g., `cmd/server/main.go`).
*   `internal/`: Houses the core logic of the application, including API handlers, authentication, database interactions, GitHub client, and the syncer service.
*   `migrations/`: Database migration files.
*   `web/frontend/`: Contains the Vue.js frontend application.
*   `docker-compose.yml`: Defines the services, networks, and volumes for Docker.
*   `Dockerfile`: Instructions to build the Docker image for the application.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.
