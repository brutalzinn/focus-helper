# Focus Helper - Development Guide

This guide helps developers get started with Focus Helper development on any machine.

## 🚀 Quick Start

### First Time Setup

```bash
# Clone the repository
git clone https://github.com/robertocpaes/focus-helper.git
cd focus-helper

# Complete setup and run (one command!)
make quick-start
```

That's it! The `make quick-start` command will:
- Check system requirements
- Install all dependencies
- Set up the development environment
- Build the application
- Run tests
- Start the application

### Daily Development

```bash
# Run in development mode
make dev-run

# Run tests
make dev-test

# Clean up when done
make dev-clean
```

## 📋 System Requirements

### Minimum Requirements
- **Go**: 1.21 or later
- **CMake**: 3.15 or later
- **Make**: 4.0 or later
- **Git**: Any recent version
- **Python**: 3.8 or later (for voice features)

### Supported Operating Systems
- **Linux**: Ubuntu, Debian, CentOS, RHEL, Fedora
- **macOS**: 10.15 or later
- **Windows**: 10 or later (with WSL2 recommended)

## 🛠️ Development Commands

### Setup Commands
```bash
make dev-setup      # Complete development setup
make install-deps   # Install system dependencies
make check-deps     # Check system requirements
```

### Build Commands
```bash
make build          # Build production binary
make build-debug    # Build debug binary
make run            # Run in development mode
```

### Testing Commands
```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make test-package PACKAGE=src/pkg/actions  # Test specific package
```

### Development Commands
```bash
make dev-run        # Run in development mode
make dev-test       # Run tests with coverage
make dev-clean      # Clean development artifacts
```

### Voice Commands
```bash
make download-voices  # Download voice models
```

### Documentation Commands
```bash
make docs-serve     # Start documentation server
make docs-build     # Build documentation
```

### Docker Commands
```bash
make docker-build   # Build Docker image
make docker-run     # Run Docker container
```

## 🏗️ Project Structure

```
focus-helper/
├── src/                    # Source code
│   ├── cmd/               # Application entry points
│   │   └── focus-helper/  # Main application
│   └── pkg/               # Packages
│       ├── actions/       # Action execution
│       ├── activity/      # Activity monitoring
│       ├── audio/         # Audio processing
│       ├── config/        # Configuration management
│       ├── database/      # Database operations
│       ├── integrations/  # External integrations
│       ├── language/      # Internationalization
│       ├── llm/           # LLM integration
│       ├── mcp/           # MCP server
│       ├── models/        # Data models
│       ├── notifications/ # Notification system
│       ├── persona/       # AI personas
│       ├── server/        # Web server
│       ├── state/         # Application state
│       ├── utils/         # Utilities
│       ├── variables/     # Variable management
│       └── voice/         # Voice processing
├── docs/                  # Documentation
├── scripts/               # Development scripts
├── assets/                # Static assets
├── voices/                # Voice models
├── whisper.cpp/           # Whisper submodule
├── Makefile              # Build system
└── profiles.json         # Configuration
```

## 🔧 Configuration

### Development Configuration
The setup script creates `profiles-dev.json` for development:

```json
{
  "name": "development",
  "debug": true,
  "log_level": "debug",
  "voice": {
    "enabled": false
  },
  "mcp": {
    "enabled": true,
    "port": 8089
  },
  "webhook": {
    "enabled": false
  },
  "llm": {
    "enabled": false
  }
}
```

### Environment Variables
You can use environment variables in configuration:

```bash
export FOCUSHELPER_DEBUG=true
export FOCUSHELPER_MCP_PORT=8089
export FOCUSHELPER_VOICE_ENABLED=false
```

## 🧪 Testing

### Running Tests
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
make test-package PACKAGE=src/pkg/actions

# Run integration tests (requires external services)
go test -v -tags=integration ./src/pkg/integrations/...
```

### Test Structure
- **Unit Tests**: Fast, isolated tests in `*_test.go` files
- **Integration Tests**: Tests requiring external services (tagged with `integration`)
- **Coverage**: HTML report generated in `coverage.html`

## 🎤 Voice Development

### Downloading Voice Models
```bash
make download-voices
```

### Voice Configuration
```json
{
  "voice": {
    "enabled": true,
    "engine": "onnx",
    "model_path": "./voices/pt_BR-cadu-medium.onnx",
    "model_config": "./voices/pt_BR-cadu-medium.onnx.json",
    "speaker_id": "default",
    "language": "pt-BR"
  }
}
```

## 🔗 Integration Development

### n8n Integration
See `docs/content/integrations/n8n-workflows.md` for detailed integration examples.

### MCP Server
The MCP server runs on port 8089 by default and provides:
- Session information
- Hyperfocus status
- Alert triggering
- State management

### Webhook Integration
Configure webhooks in your profile:
```json
{
  "webhook": {
    "enabled": true,
    "url": "http://your-webhook-url.com/endpoint"
  }
}
```

## 🐳 Docker Development

### Building Docker Image
```bash
make docker-build
```

### Running in Docker
```bash
make docker-run
```

### Docker Compose
```bash
docker-compose up -d
```

## 📚 Documentation

### Local Documentation Server
```bash
make docs-serve
```
Then visit `http://localhost:1313`

### Building Documentation
```bash
make docs-build
```

## 🔍 Debugging

### Debug Build
```bash
make build-debug
./focushelper-debug -debug
```

### Logging
- **Debug Mode**: `-debug` flag enables verbose logging
- **Log Files**: `focus_helper.log` and `focus_helper_debug.log`
- **Log Levels**: `debug`, `info`, `warn`, `error`

### Common Issues

#### Build Issues
```bash
# Clean and rebuild
make clean
make build

# Check dependencies
make check-deps
```

#### Voice Issues
```bash
# Download voices
make download-voices

# Check audio system
arecord -l  # List audio devices
```

#### Permission Issues
```bash
# Fix script permissions
chmod +x scripts/*.sh

# Fix binary permissions
chmod +x focushelper
```

## 🚀 Contributing

### Development Workflow
1. **Fork** the repository
2. **Clone** your fork
3. **Setup** development environment: `make quick-start`
4. **Create** a feature branch
5. **Make** your changes
6. **Test** your changes: `make dev-test`
7. **Commit** your changes
8. **Push** to your fork
9. **Create** a pull request

### Code Style
- Follow Go conventions
- Use `gofmt` and `go vet`
- Write tests for new features
- Update documentation

### Git Hooks
Pre-commit hooks are automatically installed and will:
- Run `go fmt`
- Run `go vet`
- Run tests

## 📞 Getting Help

### Resources
- **Documentation**: `docs/` directory
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Community**: Discord/Slack (if available)

### Common Commands Reference
```bash
# Quick reference
make help                    # Show all commands
make quick-start            # First time setup
make dev-run               # Daily development
make dev-test              # Run tests
make dev-clean             # Clean up
make download-voices       # Setup voices
make docs-serve            # View documentation
```

## 🎯 Next Steps

After setup, you can:
1. **Explore** the codebase in `src/`
2. **Read** the documentation in `docs/`
3. **Test** voice features with `make download-voices`
4. **Set up** n8n integration
5. **Contribute** to the project

Happy coding! 🚀
