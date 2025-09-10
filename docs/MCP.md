# MCP Server for Focus Helper

The Focus Helper application includes a Model Context Protocol (MCP) server that exposes session information and allows triggering specific alert levels. This enables integration with other applications and AI systems.

## Features

- **Session Information**: Get current session details, duration, and status
- **Hyperfocus Monitoring**: Monitor hyperfocus levels and duration
- **Alert Level Management**: View available alert levels and trigger them
- **Real-time Data**: Access live session data via HTTP API
- **JSON-RPC Protocol**: Standardized communication protocol

## API Endpoints

### Base URL
```
http://localhost:8089
```

### Health Check
```
GET /health
```

### MCP Endpoints
```
POST /mcp
```

## Available Methods

### 1. Get Session Information
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "get_session_info",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "session_id": "session_123",
    "subject": "General Use",
    "start_time": "2025-09-09T18:00:00Z",
    "current_time": "2025-09-09T18:30:00Z",
    "duration": "30m0s",
    "is_active": true,
    "hyperfocus_level": "low",
    "hyperfocus_start_time": "2025-09-09T18:15:00Z",
    "hyperfocus_duration": "15m0s",
    "last_activity_time": "2025-09-09T18:29:30Z",
    "idle_duration": "30s"
  }
}
```

### 2. Get Alert Levels
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "get_alert_levels",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": [
    {
      "index": 0,
      "level": "low",
      "enabled": true,
      "threshold": "45m0s",
      "tolerance": 1.0
    },
    {
      "index": 1,
      "level": "medium",
      "enabled": true,
      "threshold": "90m0s",
      "tolerance": 0.8
    }
  ]
}
```

### 3. Trigger Alert
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "trigger_alert",
  "params": {
    "alert_index": 0
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "success": true,
    "message": "Alert level low triggered",
    "level": "low"
  }
}
```

### 4. Get Hyperfocus Status
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "get_hyperfocus_status",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "is_hyperfocus": true,
    "level": "low",
    "start_time": "2025-09-09T18:15:00Z",
    "duration": "15m0s"
  }
}
```

### 5. Ping
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "ping",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "pong": "focus-helper-mcp"
  }
}
```

## Configuration

### Command Line Flags
```bash
# Enable MCP server
focushelper --mcp

# Enable MCP server with custom port
focushelper --mcp --mcp-port 9090

# Enable MCP server in Docker mode
focushelper --docker --mcp
```

### Configuration File
```json
{
  "mcp_server_enabled": true,
  "mcp_server_port": 8089
}
```

## Usage Examples

### Using curl

```bash
# Get session information
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "get_session_info",
    "params": {}
  }'

# Trigger alert level 0
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "trigger_alert",
    "params": {
      "alert_index": 0
    }
  }'
```

### Using the Go Client

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    client := NewMCPClient("http://localhost:8089")
    
    // Get session info
    sessionInfo, err := client.GetSessionInfo()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Session duration: %s\n", sessionInfo.Duration)
    
    // Trigger alert
    if err := client.TriggerAlert(0); err != nil {
        panic(err)
    }
}
```

### Using Python

```python
import requests
import json

def get_session_info():
    url = "http://localhost:8089/mcp"
    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "get_session_info",
        "params": {}
    }
    
    response = requests.post(url, json=payload)
    return response.json()

def trigger_alert(alert_index):
    url = "http://localhost:8089/mcp"
    payload = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "trigger_alert",
        "params": {
            "alert_index": alert_index
        }
    }
    
    response = requests.post(url, json=payload)
    return response.json()

# Usage
session = get_session_info()
print(f"Session duration: {session['result']['duration']}")

result = trigger_alert(0)
print(f"Alert triggered: {result['result']['success']}")
```

## Error Handling

The MCP server returns standard JSON-RPC error responses:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "alert_index parameter required"
  }
}
```

### Error Codes
- `-32700`: Parse error
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

## Security Considerations

- The MCP server binds to `127.0.0.1` by default for local access only
- CORS headers are set to allow cross-origin requests
- No authentication is implemented (suitable for local use)
- For production use, consider adding authentication and HTTPS

## Integration Examples

### With AI Assistants
The MCP server can be integrated with AI assistants to provide context about focus sessions:

```python
# Example integration with an AI assistant
def check_focus_status():
    session = get_session_info()
    if session['result']['is_active']:
        duration = session['result']['duration']
        if duration > timedelta(hours=2):
            trigger_alert(2)  # High alert
            return "You've been focused for over 2 hours. Consider taking a break."
    return "Focus session is normal."
```

### With Monitoring Tools
```bash
# Monitor session duration
watch -n 30 'curl -s -X POST http://localhost:8089/mcp -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"get_session_info\",\"params\":{}}" | jq ".result.duration"'
```

### With Automation Scripts
```bash
#!/bin/bash
# Auto-trigger break reminder after 1 hour
while true; do
    DURATION=$(curl -s -X POST http://localhost:8089/mcp \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"get_session_info","params":{}}' \
        | jq -r '.result.duration' | sed 's/m.*//')
    
    if [ "$DURATION" -gt 60 ]; then
        curl -s -X POST http://localhost:8089/mcp \
            -H "Content-Type: application/json" \
            -d '{"jsonrpc":"2.0","id":2,"method":"trigger_alert","params":{"alert_index":0}}'
        break
    fi
    
    sleep 300  # Check every 5 minutes
done
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure the MCP server is enabled in configuration
   - Check if the port is available
   - Verify the application is running

2. **Method Not Found**
   - Check the method name spelling
   - Ensure you're using the correct JSON-RPC format

3. **Invalid Params**
   - Verify parameter names and types
   - Check required parameters are included

### Debug Mode
Enable debug logging to see MCP server activity:

```bash
focushelper --debug --mcp
```

### Health Check
Test if the MCP server is running:

```bash
curl http://localhost:8089/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "focus-helper-mcp"
}
```
