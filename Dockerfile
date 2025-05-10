# Build stage
FROM golang:1.22-alpine AS builder

# Install git for fetching dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
# Note: .env files are excluded via .dockerignore
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o ai-dev-env ./cmd/ai-dev-env

# Final stage
FROM alpine:latest

# Add ca-certificates for HTTPS requests to OpenRouter API
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/ai-dev-env /app/ai-dev-env

# Copy the web directory
COPY --from=builder /app/web /app/web

# Copy the .env.template file for reference
COPY --from=builder /app/.env.template /app/.env.template

# Expose the port
EXPOSE 8080

# Environment variables are now provided through docker-compose.yaml or at runtime
# The OPENROUTER_API_KEY should be provided via:
# 1. docker-compose environment section
# 2. docker-compose env_file section
# 3. docker run -e OPENROUTER_API_KEY=your_key
# 4. web UI input

# Run the binary
CMD ["/app/ai-dev-env"] 