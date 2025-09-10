# Multi-stage build for focus-helper
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    g++ \
    make \
    cmake \
    pkgconfig \
    alsa-lib-dev \
    portaudio-dev \
    ffmpeg-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    alsa-lib \
    portaudio \
    ffmpeg \
    sox \
    pulseaudio-utils

# Create app user
RUN adduser -D -s /bin/sh focushelper

# Set working directory
WORKDIR /app

# Copy built binary
COPY --from=builder /app/focushelper /usr/local/bin/focushelper

# Copy configuration files
COPY --from=builder /app/profiles.json /app/profiles.json
COPY --from=builder /app/langs /app/langs
COPY --from=builder /app/assets /app/assets
COPY --from=builder /app/voices /app/voices

# Create config directory
RUN mkdir -p /home/focushelper/.config/focushelper

# Copy configuration to user directory
RUN cp -r /app/* /home/focushelper/.config/focushelper/

# Set ownership
RUN chown -R focushelper:focushelper /home/focushelper

# Switch to app user
USER focushelper

# Set environment variables for Docker mode
ENV FOCUSHELPER_DOCKER_MODE=true

# Expose port
EXPOSE 8088

# Run the application in Docker mode
CMD ["focushelper", "--docker"]