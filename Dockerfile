# Shelly CLI Docker Image
# Used by GoReleaser - expects pre-built binary in context

FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -u 1000 shelly
USER shelly

# Copy pre-built binary from GoReleaser
COPY shelly /usr/local/bin/shelly

# Set up working directory
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
