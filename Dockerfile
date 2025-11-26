# Use official Go image for building
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git (required for Go modules)
RUN apk add --no-cache git

# Copy go.mod + go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY . .

# Build the server binary
RUN go build -o mini-hotel ./cmd/server

# Final image
FROM alpine:latest
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mini-hotel .

# Expose port
EXPOSE 8080

# Default command
CMD ["./mini-hotel"]
