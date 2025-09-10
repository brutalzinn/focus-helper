# Focus Helper

**AI-powered focus management for autistic individuals**

Focus Helper is a sophisticated application designed to help autistic individuals manage their hyperfocus experiences. When hyperfocus becomes overwhelming and requires external intervention, Focus Helper provides intelligent monitoring, voice commands, and automated alerts to help maintain healthy focus patterns.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/robertocpaes/focus-helper.git
cd focus-helper

# Build and install
make build
make move

# Start the application
focushelper --mcp
```

## 📚 Documentation

**📖 [View Full Documentation](docs/)**

Our comprehensive documentation is built with Hugo and includes:

- **[Getting Started](docs/content/getting-started/)** - Quick setup and first run
- **[Features](docs/content/features/)** - Complete feature overview
- **[MCP Integration](docs/content/mcp-integration/)** - External tool integration
- **[Docker Support](docs/content/docker/)** - Containerized deployment
- **[Examples](docs/content/examples/)** - Practical examples and use cases
  - [n8n Integration](docs/content/examples/n8n-integration.md) - Workflow automation
  - [Cursor IDE Integration](docs/CURSOR_INTEGRATION.md) - Code editor integration
  - [Go Client](docs/content/examples/mcp_client.go) - Go client library
  - [Python Client](docs/content/examples/mcp_client.py) - Python client library

### Local Documentation

To view the documentation locally:

```bash
cd docs
make setup    # First time setup
make serve    # Start development server
```

Then open `http://localhost:1313` in your browser.

## ✨ Key Features

- **🧠 Intelligent Hyperfocus Detection** - AI-powered analysis with configurable alert levels and optimized LLM calls
- **🎤 Voice Commands** - Natural language voice recognition in Portuguese and English
- **⏱️ Session Management** - Persistent session tracking with configurable timeouts
- **🔌 MCP Integration** - Model Context Protocol server for external tool integration
- **🐳 Docker Support** - Containerized deployment with flexible audio handling
- **📊 Real-time Monitoring** - Live activity tracking and analytics
- **⚡ Performance Optimized** - Configurable LLM call intervals to reduce computational overhead

## 🛠️ Installation

### Prerequisites

- **Go 1.21+** (for building from source)
- **CMake** (for building Whisper.cpp)
- **PortAudio** (for audio support)
- **Git** (for cloning the repository)

### Build from Source

```bash
git clone https://github.com/robertocpaes/focus-helper.git
cd focus-helper
make build
make move
```

### Docker

```bash
docker-compose up --build
```

## 🎯 Use Cases

### For Autistic Individuals
- **Hyperfocus Management**: Get alerts when focus becomes unhealthy
- **Break Reminders**: Automated reminders to take breaks
- **Activity Tracking**: Monitor computer usage patterns
- **Voice Control**: Hands-free interaction with the system

### For Caregivers
- **Remote Monitoring**: Track focus sessions via MCP API
- **Alert Management**: Receive notifications about concerning patterns
- **Session Analytics**: Understand focus patterns over time
- **Custom Interventions**: Trigger specific alert levels remotely

### For Developers
- **MCP Integration**: Build tools that interact with focus sessions
- **API Access**: Real-time session data and control
- **Custom Alerts**: Create specialized alert systems
- **Data Analysis**: Access session data for research

## 🔧 Configuration

Focus Helper stores configuration in:
- **Linux/macOS**: `~/.config/focushelper/`
- **Windows**: `%APPDATA%\focushelper\`

Edit `profiles.json` to customize alert levels, voice commands, and other settings.

## 🐳 Docker Support

```bash
# Basic usage
docker-compose up

# With custom configuration
docker-compose -f docker-compose.prod.yml up
```

See [Docker Documentation](docs/content/docker/) for detailed setup instructions.

## 🔌 API Integration

Focus Helper includes a comprehensive MCP server for external integration:

```bash
# Start with MCP server
focushelper --mcp

# Test the API
curl http://localhost:8089/health
```

### Example API Usage

```bash
# Get session information
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"get_session_info","params":{}}'

# Trigger alert
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"trigger_alert","params":{"alert_index":0}}'
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Whisper.cpp** for voice recognition
- **PortAudio** for cross-platform audio support
- **Hugo** for documentation
- **The autistic community** for inspiration and feedback

## 📞 Support

- 📖 **Documentation**: [docs/](docs/)
- 🐛 **Issues**: [GitHub Issues](https://github.com/robertocpaes/focus-helper/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/robertocpaes/focus-helper/discussions)
- 📧 **Email**: [Contact us](mailto:support@focus-helper.dev)

---

**Ready to get started?** Check out our [Getting Started Guide](docs/content/getting-started/) or explore the [Features](docs/content/features/) section to learn more about what Focus Helper can do for you.
