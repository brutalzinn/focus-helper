---
title: "Getting Started"
description: "Quick start guide for Focus Helper"
date: 2025-09-09T19:00:00Z
draft: false
weight: 10
---

# Getting Started with Focus Helper

This guide will help you get Focus Helper up and running quickly on your system.

## Prerequisites

Before installing Focus Helper, ensure you have:

- **Go 1.21+** (for building from source)
- **CMake** (for building Whisper.cpp)
- **PortAudio** (for audio support)
- **ALSA/OSS** (Linux audio system)
- **Git** (for cloning the repository)

### Installing Dependencies

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install -y golang-go cmake build-essential portaudio19-dev libasound2-dev
```

#### macOS
```bash
# Install Homebrew if you haven't already
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install go cmake portaudio
```

#### Windows
```bash
# Install Chocolatey if you haven't already
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install dependencies
choco install golang cmake portaudio
```

## Installation

### Option 1: Build from Source (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/robertocpaes/focus-helper.git
   cd focus-helper
   ```

2. **Build the application**:
   ```bash
   make build
   ```

3. **Install to system**:
   ```bash
   make move
   ```

### Option 2: Docker (Alternative)

1. **Build Docker image**:
   ```bash
   docker build -t focus-helper .
   ```

2. **Run with Docker Compose**:
   ```bash
   docker-compose up --build
   ```

## First Run

### Basic Usage

Start Focus Helper with default settings:

```bash
focushelper
```

### With MCP Server

Enable the MCP server for external integration:

```bash
focushelper --mcp
```

### Docker Mode

Run in Docker-compatible mode:

```bash
focushelper --docker
```

### Microphone Selection

Force microphone selection on startup:

```bash
focushelper --select-microphone
```

## Initial Configuration

### 1. Microphone Setup

On first run, Focus Helper will prompt you to select a microphone:

```
🎤 Microphone Selection
Available input devices:
[0] HDA Intel PCH: ALC1220 Analog (Input Channels: 2, DefaultSampleRate: 48000)
[1] USB Audio Device (Input Channels: 1, DefaultSampleRate: 44100)
Select input device by number: 0
```

### 2. Voice Commands

Focus Helper comes with pre-configured voice commands:

- **Activation**: "torre controle comando" (Portuguese)
- **Emergency**: "mayday emergencia"
- **Stop**: "parar para cancela"
- **Status**: "tempo horas atividade"
- **Check**: "check checagem status"

### 3. Alert Levels

Default alert levels are configured as:

| Level | Duration | Description |
|-------|----------|-------------|
| Low | 45 minutes | Basic break reminder |
| Medium | 90 minutes | More urgent reminder |
| High | 3 hours | Strong intervention needed |
| Critical | 6 hours | Emergency intervention |

## Configuration Files

Focus Helper stores configuration in:

- **Linux/macOS**: `~/.config/focushelper/`
- **Windows**: `%APPDATA%\focushelper\`

### Main Configuration

Edit `profiles.json` to customize:

```json
{
  "default": {
    "alert_levels": [
      {
        "enabled": true,
        "level": "low",
        "threshold": "45m",
        "actions": [
          {
            "type": "sound",
            "sound_file": "alert_level_1.mp3"
          }
        ]
      }
    ],
    "mcp_server_enabled": true,
    "mcp_server_port": 8089
  }
}
```

## Testing the Installation

### 1. Check Health Status

```bash
curl http://localhost:8088/health
```

Expected response:
```json
{"status": "healthy", "service": "focus-helper"}
```

### 2. Test MCP Server

```bash
curl http://localhost:8089/health
```

Expected response:
```json
{"status": "healthy", "service": "focus-helper-mcp"}
```

### 3. Test Voice Commands

1. Say the activation phrase: "torre controle comando"
2. Wait for the system to respond
3. Try a command like "tempo horas atividade"

### 4. Test Alert Triggering

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"trigger_alert","params":{"alert_index":0}}'
```

## Troubleshooting

### Common Issues

#### Audio Not Working
```bash
# Check audio devices
arecord -l

# Test microphone
arecord -f cd -d 5 test.wav
aplay test.wav
```

#### Permission Denied
```bash
# Add user to audio group (Linux)
sudo usermod -a -G audio $USER
# Log out and back in
```

#### Port Already in Use
```bash
# Check what's using the port
sudo netstat -tulpn | grep :8088
sudo netstat -tulpn | grep :8089

# Kill the process
sudo kill -9 <PID>
```

#### Build Errors
```bash
# Clean and rebuild
make clean
make build
```

### Getting Help

- 📖 **Documentation**: Check the [Features](/features/) and [Configuration](/configuration/) sections
- 🐛 **Issues**: Report problems on [GitHub Issues](https://github.com/robertocpaes/focus-helper/issues)
- 💬 **Discussions**: Ask questions in [GitHub Discussions](https://github.com/robertocpaes/focus-helper/discussions)

## Next Steps

Now that you have Focus Helper installed and running:

1. **Explore Features**: Learn about [hyperfocus detection](/features/hyperfocus-detection/) and [voice commands](/features/voice-commands/)
2. **Configure Settings**: Customize [alert levels](/configuration/alert-levels/) and [voice commands](/configuration/voice-commands/)
3. **Integrate Tools**: Set up [MCP integration](/mcp-integration/) with your favorite tools
4. **Docker Deployment**: Learn about [Docker setup](/docker/) for production use

---

**Ready to dive deeper?** Check out our [Features](/features/) section to learn about all the capabilities of Focus Helper!
