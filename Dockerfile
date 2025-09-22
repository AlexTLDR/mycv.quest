# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates, assets, and utils directories
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/utils ./utils

# Change ownership to non-root user
RUN chown -R appuser:appgroup /root

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 7070

# Run the binary
CMD ["./main", "-serve", "-port=7070"]
