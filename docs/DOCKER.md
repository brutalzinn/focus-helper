# Docker Support for Focus Helper

This document explains how to run Focus Helper in Docker containers for better portability and isolation.

## Features

- **Docker Mode**: Automatically detects and uses the first available audio device
- **Sample Rate Fallback**: Tries multiple sample rates (16kHz, 44.1kHz, 48kHz, etc.) for compatibility
- **Audio Device Access**: Properly configured for audio input/output in containers
- **Persistent Configuration**: Configuration and data are persisted in Docker volumes

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Build and run
docker-compose up --build

# Run in background
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Using Docker directly

```bash
# Build the image
docker build -t focus-helper .

# Run the container
docker run -d \
  --name focus-helper \
  --device /dev/snd:/dev/snd \
  --privileged \
  --group-add audio \
  -p 8088:8088 \
  focus-helper
```

## Configuration

### Environment Variables

- `FOCUSHELPER_DOCKER_MODE=true`: Enables Docker-compatible mode
- `OLLAMA_ENDPOINT`: Optional Ollama server endpoint for AI features

### Volume Mounts

- `focus-helper-data:/home/focushelper/.config/focushelper`: Persistent configuration and data

## Audio Requirements

The container needs access to audio devices:

- `--device /dev/snd:/dev/snd`: Audio device access
- `--privileged`: Required for audio device access
- `--group-add audio`: Audio group membership

## Troubleshooting

### Audio Issues

If you encounter audio problems:

1. **Check audio device access**:
   ```bash
   docker exec -it focus-helper ls -la /dev/snd/
   ```

2. **Test audio device detection**:
   ```bash
   docker exec -it focus-helper focushelper --docker --select-microphone
   ```

3. **Check logs for sample rate issues**:
   ```bash
   docker logs focus-helper | grep "sample rate"
   ```

### Common Issues

- **"No input devices found"**: Audio device not properly mounted
- **"Invalid sample rate"**: Audio device doesn't support required sample rates (handled automatically)
- **Permission denied**: Missing `--privileged` flag or audio group

## Development

### Building for Development

```bash
# Build without cache
docker-compose build --no-cache

# Run with volume mounts for development
docker-compose -f docker-compose.dev.yml up
```

### Testing Audio

```bash
# Test microphone selection
docker exec -it focus-helper focushelper --select-microphone

# Test Docker mode
docker exec -it focus-helper focushelper --docker
```

## Production Deployment

For production deployment:

1. **Use specific image tags** instead of `latest`
2. **Set resource limits** in docker-compose.yml
3. **Configure logging** with proper log drivers
4. **Set up monitoring** for the application
5. **Use secrets** for sensitive configuration

Example production docker-compose.yml:

```yaml
services:
  focus-helper:
    image: focus-helper:v1.0.0
    container_name: focus-helper
    restart: unless-stopped
    ports:
      - "8088:8088"
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
```
