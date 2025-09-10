---
title: "Docker Support"
description: "Containerized deployment with Docker"
date: 2025-09-09T19:00:00Z
draft: false
weight: 40
---

# Docker Support

Focus Helper provides comprehensive Docker support for easy deployment and cross-platform compatibility.

## Overview

Docker support enables:
- **Cross-Platform Deployment**: Run on any Docker-supported platform
- **Easy Setup**: One-command deployment
- **Audio Support**: Full audio device access in containers
- **Flexible Configuration**: Environment-based configuration
- **Production Ready**: Scalable containerized deployment

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/robertocpaes/focus-helper.git
   cd focus-helper
   ```

2. **Start with Docker Compose**:
   ```bash
   docker-compose up --build
   ```

3. **Run in background**:
   ```bash
   docker-compose up -d --build
   ```

### Using Docker Directly

1. **Build the image**:
   ```bash
   docker build -t focus-helper .
   ```

2. **Run the container**:
   ```bash
   docker run -d \
     --name focus-helper \
     --device /dev/snd:/dev/snd \
     --privileged \
     --group-add audio \
     -p 8088:8088 \
     -p 8089:8089 \
     focus-helper
   ```

## Docker Compose Configuration

### Basic Configuration

```yaml
version: '3.8'

services:
  focus-helper:
    build: .
    container_name: focus-helper
    restart: unless-stopped
    ports:
      - "8088:8088"  # Web interface
      - "8089:8089"  # MCP server
    environment:
      - FOCUSHELPER_DOCKER_MODE=true
    volumes:
      - focus-helper-data:/home/focushelper/.config/focushelper
    devices:
      - "/dev/snd:/dev/snd"
    privileged: true
    cap_add:
      - SYS_PTRACE
    security_opt:
      - seccomp:unconfined
    group_add:
      - audio

volumes:
  focus-helper-data:
```

### Advanced Configuration

```yaml
version: '3.8'

services:
  focus-helper:
    build: .
    container_name: focus-helper
    restart: unless-stopped
    ports:
      - "8088:8088"
      - "8089:8089"
    environment:
      - FOCUSHELPER_DOCKER_MODE=true
      - FOCUSHELPER_MCP_SERVER_ENABLED=true
      - FOCUSHELPER_MCP_SERVER_PORT=8089
      - FOCUSHELPER_DEBUG=false
    volumes:
      - focus-helper-data:/home/focushelper/.config/focushelper
      - ./custom-profiles.json:/home/focushelper/.config/focushelper/profiles.json:ro
    devices:
      - "/dev/snd:/dev/snd"
    privileged: true
    cap_add:
      - SYS_PTRACE
    security_opt:
      - seccomp:unconfined
    group_add:
      - audio
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  focus-helper-data:
```

## Dockerfile

### Multi-Stage Build

```dockerfile
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

# Expose ports
EXPOSE 8088 8089

# Run the application in Docker mode
CMD ["focushelper", "--docker"]
```

## Audio Configuration

### Linux Audio Support

For Linux systems, ensure proper audio device access:

```yaml
services:
  focus-helper:
    # ... other configuration
    devices:
      - "/dev/snd:/dev/snd"
    privileged: true
    group_add:
      - audio
```

### macOS Audio Support

For macOS, use PulseAudio:

```yaml
services:
  focus-helper:
    # ... other configuration
    environment:
      - OTO_DRIVER=pulseaudio
      - PULSE_SERVER=${XDG_RUNTIME_DIR}/pulse/native
    volumes:
      - ${XDG_RUNTIME_DIR}/pulse/native:${XDG_RUNTIME_DIR}/pulse/native
```

### Windows Audio Support

For Windows, use DirectSound:

```yaml
services:
  focus-helper:
    # ... other configuration
    environment:
      - OTO_DRIVER=directsound
```

## Environment Variables

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FOCUSHELPER_DOCKER_MODE` | `false` | Enable Docker-compatible mode |
| `FOCUSHELPER_MCP_SERVER_ENABLED` | `true` | Enable MCP server |
| `FOCUSHELPER_MCP_SERVER_PORT` | `8089` | MCP server port |
| `FOCUSHELPER_DEBUG` | `false` | Enable debug mode |

### Audio Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OTO_DRIVER` | `auto` | Audio driver to use |
| `PULSE_SERVER` | - | PulseAudio server (macOS) |
| `ALSA_DEVICE` | - | ALSA device (Linux) |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FOCUSHELPER_DATABASE_FILE` | `/home/focushelper/.config/focushelper/focus_helper.db` | Database file path |
| `FOCUSHELPER_LOG_FILE` | `/home/focushelper/.config/focushelper/focus_helper.log` | Log file path |

## Volume Management

### Persistent Data

```yaml
volumes:
  focus-helper-data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /path/to/persistent/data
```

### Configuration Files

```yaml
services:
  focus-helper:
    volumes:
      - ./profiles.json:/home/focushelper/.config/focushelper/profiles.json:ro
      - ./custom-langs:/home/focushelper/.config/focushelper/langs:ro
      - ./custom-assets:/home/focushelper/.config/focushelper/assets:ro
```

## Networking

### Port Configuration

```yaml
services:
  focus-helper:
    ports:
      - "8088:8088"  # Web interface
      - "8089:8089"  # MCP server
      - "8080:8080"  # Custom port
```

### Network Mode

```yaml
services:
  focus-helper:
    network_mode: host  # Use host networking
    # OR
    networks:
      - focus-helper-network

networks:
  focus-helper-network:
    driver: bridge
```

## Security Considerations

### Privileged Mode

```yaml
services:
  focus-helper:
    privileged: true  # Required for audio device access
    cap_add:
      - SYS_PTRACE
    security_opt:
      - seccomp:unconfined
```

### User Permissions

```yaml
services:
  focus-helper:
    user: "1000:1000"  # Run as specific user
    group_add:
      - audio
      - video
```

## Monitoring and Logging

### Log Configuration

```yaml
services:
  focus-helper:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Health Checks

```yaml
services:
  focus-helper:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8088/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Production Deployment

### Resource Limits

```yaml
services:
  focus-helper:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

### Restart Policies

```yaml
services:
  focus-helper:
    restart: unless-stopped
    # OR
    restart: on-failure:3
```

### Scaling

```yaml
services:
  focus-helper:
    deploy:
      replicas: 1
      update_config:
        parallelism: 1
        delay: 10s
      rollback_config:
        parallelism: 1
        delay: 10s
```

## Troubleshooting

### Common Issues

#### Audio Not Working
```bash
# Check audio devices
docker exec -it focus-helper ls -la /dev/snd/

# Test audio device detection
docker exec -it focus-helper focushelper --docker --select-microphone
```

#### Permission Denied
```bash
# Check container permissions
docker exec -it focus-helper id

# Check audio group membership
docker exec -it focus-helper groups
```

#### Port Already in Use
```bash
# Check port usage
sudo netstat -tulpn | grep :8088
sudo netstat -tulpn | grep :8089

# Kill conflicting processes
sudo kill -9 <PID>
```

### Debug Mode

```bash
# Run with debug logging
docker run -it --rm \
  --device /dev/snd:/dev/snd \
  --privileged \
  --group-add audio \
  -e FOCUSHELPER_DEBUG=true \
  focus-helper focushelper --docker --debug
```

### Logs

```bash
# View container logs
docker logs focus-helper

# Follow logs in real-time
docker logs -f focus-helper

# View specific log files
docker exec -it focus-helper tail -f /home/focushelper/.config/focushelper/focus_helper.log
```

## Best Practices

1. **Use Docker Compose**: Easier management and configuration
2. **Persistent Volumes**: Store configuration and data persistently
3. **Resource Limits**: Set appropriate memory and CPU limits
4. **Health Checks**: Monitor container health
5. **Logging**: Configure proper log rotation
6. **Security**: Use least privilege principle
7. **Updates**: Regular image updates for security

---

**Ready to deploy?** Check out our [Configuration](/configuration/) section for detailed configuration options and [Examples](/examples/) for deployment scenarios.
