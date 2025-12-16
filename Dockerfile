# Shelly CLI Docker Image
# Used by GoReleaser dockers_v2 - expects platform-specific binary

FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -u 1000 shelly
USER shelly

# Copy pre-built binary from GoReleaser (platform-specific path)
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/shelly /usr/local/bin/shelly

# Set up working directory
WORKDIR /home/shelly

# Default entrypoint
ENTRYPOINT ["shelly"]
CMD ["--help"]

# Labels set via goreleaser
