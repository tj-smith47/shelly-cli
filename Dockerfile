# Shelly CLI Docker Image
# Multi-stage build for minimal final image

# =============================================================================
# Build Stage
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
        -X github.com/tj-smith47/shelly-cli/internal/version.Version=${VERSION} \
        -X github.com/tj-smith47/shelly-cli/internal/version.Commit=${COMMIT} \
        -X github.com/tj-smith47/shelly-cli/internal/version.Date=${DATE} \
        -X github.com/tj-smith47/shelly-cli/internal/version.BuiltBy=docker" \
    -trimpath \
    -o shelly \
    ./cmd/shelly

# =============================================================================
# Final Stage
# =============================================================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -u 1000 shelly
USER shelly

# Copy binary from builder
COPY --from=builder /build/shelly /usr/local/bin/shelly

# Set up config directory
ENV HOME=/home/shelly
WORKDIR /home/shelly

# Default entrypoint
ENTRYPOINT ["shelly"]
CMD ["--help"]

# Labels for container metadata
LABEL org.opencontainers.image.title="Shelly CLI"
LABEL org.opencontainers.image.description="Command-line interface for Shelly smart home devices"
LABEL org.opencontainers.image.url="https://github.com/tj-smith47/shelly-cli"
LABEL org.opencontainers.image.source="https://github.com/tj-smith47/shelly-cli"
LABEL org.opencontainers.image.vendor="tj-smith47"
LABEL org.opencontainers.image.licenses="MIT"
