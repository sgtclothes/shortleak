# Backend Dockerfile with Multi-stage builds
FROM golang:1.24-alpine AS base

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Development stage
FROM base AS development
RUN echo "Development stage"
EXPOSE 8080
CMD ["go", "run", "main.go"]

# Test stage
FROM base AS test
RUN echo "Test stage"
# Install test dependencies if any
RUN go mod download
CMD ["go", "test", "./...", "-v", "-coverprofile=coverage.out"]

# Production build stage
FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest AS production

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration and seed files if they exist
COPY --from=builder /app/cmd ./cmd

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]