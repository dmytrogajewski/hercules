# Multi-stage build for Hercules
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO disabled for better compatibility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X gopkg.in/src-d/hercules.v10.BinaryGitHash=$(git rev-parse HEAD)" \
    -o hercules ./cmd/hercules

# Final stage
FROM gcr.io/distroless/static:nonroot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/hercules /usr/local/bin/hercules

# Copy documentation
COPY --from=builder /app/docs /usr/local/share/hercules/docs
COPY --from=builder /app/README.md /usr/local/share/hercules/
COPY --from=builder /app/config.yaml.example /etc/hercules/config.yaml.example

# Switch to non-root user
USER nonroot

# Expose default ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/hercules"]

# Default arguments
CMD ["server", "--config", "/etc/hercules/config.yaml"] 