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

# Expose the port
EXPOSE 8080

# Create environment variable for OpenRouter API key
ENV OPENROUTER_API_KEY=""

# Run the binary
CMD ["/app/ai-dev-env"] 