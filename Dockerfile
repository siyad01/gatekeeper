# ─── Stage 1: Build ───────────────────────────────────────
FROM golang:1.26.2-alpine AS builder

# Install git (needed for go modules)
RUN apk add --no-cache git

# Set working directory inside container
WORKDIR /app

# Copy dependency files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY *.go ./

# Build the binary
# CGO_ENABLED=0 → pure Go binary, no C dependencies
# GOOS=linux    → compile for Linux (the container OS)
RUN CGO_ENABLED=0 GOOS=linux go build -o gatekeeper .

# ─── Stage 2: Runtime ─────────────────────────────────────
FROM alpine:3.19

# Add CA certificates (needed for HTTPS calls to MCP servers)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the compiled binary from builder stage
COPY --from=builder /app/gatekeeper .

# Create directory for audit logs
RUN mkdir -p /app/logs

# Expose the port
EXPOSE 8080

# Health check — Docker will monitor this
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# Run the binary
CMD ["./gatekeeper"]