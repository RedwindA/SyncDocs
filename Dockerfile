# Stage 1: Build Frontend Assets
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package.json and lock file
COPY web/frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the frontend source code
COPY web/frontend ./

# Build the frontend application
# This will create a 'dist' directory with static assets
RUN npm run build


# Stage 2: Build the Go application
FROM golang:1.24-alpine AS go-builder

# Set the working directory inside the container
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download Go dependencies.
RUN go mod download

# Copy the Go source code (excluding frontend which is handled separately)
COPY cmd ./cmd
COPY internal ./internal
# Copy migrations if needed in the final image (uncomment below)
# COPY migrations ./migrations

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /syncdocs cmd/server/main.go


# Stage 3: Create the final, minimal image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the static Go binary from the go-builder stage
COPY --from=go-builder /syncdocs /app/syncdocs

# Copy the built frontend assets from the frontend-builder stage
# The assets will be served by the Go application
COPY --from=frontend-builder /app/frontend/dist /app/web/frontend/dist

# Copy migrations (optional, if needed at runtime)
# COPY --from=go-builder /app/migrations /app/migrations

# Expose the port the application runs on (should match SERVER_PORT env var)
# Defaulting to 8080 as per config default
EXPOSE 8080

# Set the entrypoint for the container
# The application will be run when the container starts.
ENTRYPOINT ["/app/syncdocs"]

# Optional: Add a non-root user for security
# RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# USER appuser
