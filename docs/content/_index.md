---
title: "Focus Helper"
description: "AI-powered focus management for autistic individuals"
date: 2025-09-09T19:00:00Z
draft: false
---

# Focus Helper

**AI-powered focus management for autistic individuals**

Focus Helper is a sophisticated application designed to help autistic individuals manage their hyperfocus experiences. When hyperfocus becomes overwhelming and requires external intervention, Focus Helper provides intelligent monitoring, voice commands, and automated alerts to help maintain healthy focus patterns.

## Key Features

### 🧠 **Intelligent Hyperfocus Detection**
- AI-powered analysis of focus patterns
- Configurable alert levels (low, medium, high, critical)
- Progressive time-based and AI-enhanced detection
- Real-time monitoring of computer activity

### 🎤 **Voice Commands**
- Natural language voice recognition
- Hands-free control and interaction
- Portuguese and English language support
- Customizable activation phrases

### ⏱️ **Session Management**
- Persistent session tracking
- Configurable session timeouts
- Activity-based session recovery
- Detailed session analytics

### 🔌 **MCP Integration**
- Model Context Protocol server
- External tool integration
- Real-time API access
- Cursor IDE integration

### 🐳 **Docker Support**
- Containerized deployment
- Flexible audio handling
- Cross-platform compatibility
- Easy setup and configuration

### 📊 **Real-time Monitoring**
- Live activity tracking
- Window title analysis
- Subject detection
- Idle time monitoring

## Quick Start

1. **Install Focus Helper**:
   ```bash
   # Clone the repository
   git clone https://github.com/robertocpaes/focus-helper.git
   cd focus-helper
   
   # Build the application
   make build
   make move
   ```

2. **Start the application**:
   ```bash
   # Basic usage
   focushelper
   
   # With MCP server
   focushelper --mcp
   
   # Docker mode
   focushelper --docker
   ```

3. **Configure your settings**:
   - Edit `profiles.json` to customize alert levels
   - Set up voice commands and activation phrases
   - Configure MCP server settings

## Use Cases

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

## Architecture

Focus Helper is built with a modular architecture:

- **Core Engine**: Go-based application with SQLite database
- **Voice Recognition**: Whisper.cpp integration for speech processing
- **AI Analysis**: LLM integration for intelligent focus detection
- **MCP Server**: JSON-RPC over HTTP for external integration
- **Audio System**: PortAudio for cross-platform audio handling
- **Web Interface**: HTTP server for monitoring and control

## Getting Help

- 📖 **Documentation**: Comprehensive guides and API reference
- 🐛 **Issues**: Report bugs and request features on GitHub
- 💬 **Discussions**: Community support and feature discussions
- 📧 **Contact**: Reach out for support or collaboration

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Ready to get started?** Check out our [Getting Started Guide](/getting-started/) or explore the [Features](/features/) section to learn more about what Focus Helper can do for you.
