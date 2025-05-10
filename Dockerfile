FROM golang:1.22.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN make build

# Start a new stage with a minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/ai-dev-env /app/ai-dev-env

# Copy web assets
COPY --from=builder /app/web /app/web

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./ai-dev-env"] 