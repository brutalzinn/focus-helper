---
title: "Features"
description: "Comprehensive overview of Focus Helper features"
date: 2025-09-09T19:00:00Z
draft: false
weight: 20
---

# Focus Helper Features

Focus Helper provides a comprehensive suite of features designed to help autistic individuals manage their hyperfocus experiences effectively.

## 🧠 Hyperfocus Detection

### AI-Powered Analysis
Focus Helper uses advanced AI algorithms to analyze your computer usage patterns and detect when you're entering unhealthy hyperfocus states.

**Key Features:**
- **Progressive Time-Based Detection**: Monitors session duration against configurable thresholds
- **AI-Enhanced Analysis**: Uses LLM integration for intelligent pattern recognition
- **Optimized LLM Calls**: Configurable intervals to reduce computational overhead (default: 5 minutes)
- **Real-time Monitoring**: Continuous analysis of your activity
- **Configurable Alert Levels**: Customize detection sensitivity and response

**Alert Levels:**
- **Low (45 minutes)**: Gentle break reminder
- **Medium (90 minutes)**: More urgent intervention
- **High (3 hours)**: Strong intervention needed
- **Critical (6 hours)**: Emergency intervention required

### Activity Monitoring
- **Window Title Analysis**: Tracks what applications you're using
- **Subject Detection**: Identifies the focus area (work, entertainment, etc.)
- **Idle Time Tracking**: Monitors periods of inactivity
- **Session Persistence**: Maintains session state across restarts

## 🎤 Voice Commands

### Natural Language Processing
Focus Helper supports natural language voice commands in multiple languages.

**Supported Languages:**
- Portuguese (Brazilian)
- English

**Voice Activation:**
- **Activation Phrase**: "torre controle comando" (Portuguese)
- **Wake Word Detection**: Continuous listening for activation
- **Noise Cancellation**: Filters background noise
- **Low Latency**: Fast response to voice commands

### Available Commands

#### Status Commands
- **"tempo horas atividade"**: Get current session duration
- **"check checagem status"**: Check system status
- **"verificacao analise"**: Run system analysis

#### Control Commands
- **"parar para cancela"**: Stop current actions
- **"ignora"**: Ignore current alert
- **"mayday emergencia"**: Emergency stop

#### Information Commands
- **"tempo"**: Get time information
- **"atividade"**: Get activity summary

## ⏱️ Session Management

### Persistent Sessions
Focus Helper maintains persistent sessions that survive application restarts.

**Features:**
- **Session Recovery**: Automatically resumes interrupted sessions
- **Configurable Timeouts**: Set maximum session duration
- **Activity-Based Tracking**: Sessions based on actual computer usage
- **Database Storage**: SQLite database for reliable data persistence

### Session Analytics
- **Duration Tracking**: Monitor how long you've been focused
- **Subject Analysis**: Track what you're working on
- **Pattern Recognition**: Identify focus patterns over time
- **Export Capabilities**: Export session data for analysis

## 🔌 MCP Integration

### Model Context Protocol Server
Focus Helper includes a full MCP server for external tool integration.

**API Endpoints:**
- **Session Information**: Get current session details
- **Alert Management**: View and trigger alert levels
- **Hyperfocus Status**: Monitor hyperfocus state
- **Real-time Data**: Live session monitoring

**Integration Examples:**
- **Cursor IDE**: Direct integration with your code editor
- **Monitoring Tools**: External monitoring and alerting
- **Automation Scripts**: Custom automation workflows
- **Data Analysis**: Export data for research

### External Tool Support
- **JSON-RPC Protocol**: Standard communication protocol
- **REST API**: HTTP-based API access
- **WebSocket Support**: Real-time data streaming
- **Authentication**: Secure access control

## 🐳 Docker Support

### Containerized Deployment
Focus Helper runs seamlessly in Docker containers.

**Features:**
- **Cross-Platform**: Works on any Docker-supported platform
- **Audio Support**: Full audio device access in containers
- **Flexible Configuration**: Environment-based configuration
- **Easy Deployment**: One-command deployment

**Docker Compose:**
```yaml
services:
  focus-helper:
    build: .
    ports:
      - "8088:8088"
      - "8089:8089"
    devices:
      - "/dev/snd:/dev/snd"
    environment:
      - FOCUSHELPER_DOCKER_MODE=true
```

## 📊 Real-time Monitoring

### Live Activity Tracking
Monitor your focus patterns in real-time.

**Monitoring Features:**
- **Activity Detection**: Track mouse and keyboard activity
- **Window Monitoring**: Monitor active applications
- **Time Tracking**: Precise time measurement
- **Visual Indicators**: Status indicators and notifications

### Web Interface
Access Focus Helper through a web browser.

**Web Features:**
- **Dashboard**: Real-time session overview
- **Configuration**: Web-based settings management
- **Analytics**: Session history and patterns
- **Control Panel**: Manual control of alerts and sessions

## 🎵 Audio System

### Advanced Audio Processing
Focus Helper includes sophisticated audio processing capabilities.

**Audio Features:**
- **Multiple Sample Rates**: Support for 16kHz, 44.1kHz, 48kHz
- **Noise Reduction**: Advanced noise filtering
- **Echo Cancellation**: Clear voice recognition
- **Volume Control**: Automatic volume management

### Sound Alerts
- **Customizable Sounds**: Use your own alert sounds
- **Volume Management**: Automatic volume adjustment
- **Audio Effects**: Sound effects for different alert levels
- **Silent Mode**: Option to disable audio alerts

## 🔧 Configuration Management

### Flexible Configuration
Focus Helper offers extensive configuration options.

**Configuration Areas:**
- **Alert Levels**: Customize detection thresholds
- **Voice Commands**: Add custom voice commands
- **Audio Settings**: Configure microphone and speakers
- **Session Settings**: Set session timeouts and behavior
- **MCP Settings**: Configure external integrations
- **LLM Optimization**: Set intervals for AI analysis calls

### Profile System
- **Multiple Profiles**: Different configurations for different contexts
- **Profile Switching**: Easy switching between configurations
- **Import/Export**: Share configurations between systems
- **Version Control**: Track configuration changes

## 🛡️ Privacy and Security

### Data Protection
Focus Helper is designed with privacy in mind.

**Privacy Features:**
- **Local Processing**: All data processed locally
- **No Cloud Dependencies**: No external data transmission
- **Encrypted Storage**: Sensitive data encrypted at rest
- **User Control**: Full control over data collection

### Security Measures
- **Local Network Only**: MCP server only accessible locally
- **Authentication**: Optional authentication for external access
- **Audit Logging**: Track all system activities
- **Secure Defaults**: Secure configuration out of the box

## 📱 Cross-Platform Support

### Operating System Support
Focus Helper works on multiple operating systems.

**Supported Platforms:**
- **Linux**: Full support with ALSA/OSS
- **macOS**: Native support with Core Audio
- **Windows**: Support with DirectSound/WASAPI
- **Docker**: Containerized deployment

### Hardware Requirements
- **Microphone**: Any USB or built-in microphone
- **Audio Output**: Speakers or headphones
- **RAM**: Minimum 512MB, recommended 1GB
- **Storage**: 100MB for application and models

---

**Ready to explore specific features?** Check out our detailed guides for [Configuration](/configuration/), [MCP Integration](/mcp-integration/), and [Docker Setup](/docker/).
