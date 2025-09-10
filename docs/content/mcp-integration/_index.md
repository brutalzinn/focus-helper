---
title: "MCP Integration"
description: "Model Context Protocol integration for external tools"
date: 2025-09-09T19:00:00Z
draft: false
weight: 30
---

# MCP Integration

Focus Helper includes a comprehensive Model Context Protocol (MCP) server that enables integration with external tools and applications, including Cursor IDE.

## Overview

The MCP server provides a standardized interface for external applications to interact with Focus Helper, enabling:

- **Real-time Session Monitoring**: Access live session data
- **Alert Management**: Trigger and monitor alert levels
- **Hyperfocus Detection**: Monitor hyperfocus states
- **Voice Command Integration**: Interact with voice commands
- **Data Export**: Access session analytics

## Quick Start

### 1. Enable MCP Server

Start Focus Helper with MCP server enabled:

```bash
# Basic MCP server
focushelper --mcp

# Custom port
focushelper --mcp --mcp-port 9090

# With Docker mode
focushelper --mcp --docker
```

### 2. Test Connection

Verify the MCP server is running:

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

### 3. Test API Endpoints

```bash
# Get session information
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"get_session_info","params":{}}'
```

## API Reference

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

Retrieve current session details including duration, hyperfocus status, and activity.

**Request:**
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

Retrieve available alert levels and their configurations.

**Request:**
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

Trigger a specific alert level by index.

**Request:**
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

Retrieve current hyperfocus state and duration.

**Request:**
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

Test connection to the MCP server.

**Request:**
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

## Cursor IDE Integration

### Setup

1. **Start Focus Helper with MCP**:
   ```bash
   focushelper --mcp
   ```

2. **Configure Cursor MCP Settings**:
   Add to your Cursor settings:
   ```json
   {
     "mcpServers": {
       "focus-helper": {
         "command": "node",
         "args": ["/path/to/focus-helper/mcp-server.js"],
         "env": {}
       }
     }
   }
   ```

3. **Restart Cursor**

### Available Tools in Cursor

Once connected, you'll have access to these tools:

- **`get_session_info`**: Get current focus session status
- **`get_alert_levels`**: View available alert levels
- **`trigger_alert`**: Trigger specific alert levels
- **`get_hyperfocus_status`**: Check hyperfocus state
- **`ping`**: Test connection

### Example Interactions

**Check Focus Status:**
```
You: "What's my current focus session status?"
Cursor: [Uses get_session_info tool]
Response: "You've been focused for 45 minutes on 'General Use'. 
          You're currently in a 'low' hyperfocus state."
```

**Request Break Reminder:**
```
You: "I need a break reminder"
Cursor: [Uses trigger_alert tool with index 0]
Response: "I've triggered a low-level alert for you."
```

## Client Libraries

### Go Client

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    client := NewMCPClient("http://localhost:8089")
    
    // Get session info
    session, err := client.GetSessionInfo()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Session duration: %s\n", session.Duration)
    
    // Trigger alert
    if err := client.TriggerAlert(0); err != nil {
        panic(err)
    }
}
```

### Python Client

```python
import requests
import json

class MCPClient:
    def __init__(self, base_url="http://localhost:8089"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})
    
    def get_session_info(self):
        payload = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "get_session_info",
            "params": {}
        }
        
        response = self.session.post(f"{self.base_url}/mcp", json=payload)
        return response.json()

# Usage
client = MCPClient()
session = client.get_session_info()
print(f"Session duration: {session['result']['duration']}")
```

### JavaScript/Node.js Client

```javascript
const http = require('http');

class MCPClient {
    constructor(baseUrl = 'http://localhost:8089') {
        this.baseUrl = baseUrl;
    }
    
    async sendRequest(method, params = {}) {
        return new Promise((resolve, reject) => {
            const data = JSON.stringify({
                jsonrpc: '2.0',
                id: Date.now(),
                method,
                params
            });
            
            const options = {
                hostname: 'localhost',
                port: 8089,
                path: '/mcp',
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Content-Length': Buffer.byteLength(data)
                }
            };
            
            const req = http.request(options, (res) => {
                let body = '';
                res.on('data', chunk => body += chunk);
                res.on('end', () => {
                    try {
                        resolve(JSON.parse(body));
                    } catch (e) {
                        reject(e);
                    }
                });
            });
            
            req.on('error', reject);
            req.write(data);
            req.end();
        });
    }
    
    async getSessionInfo() {
        return this.sendRequest('get_session_info');
    }
    
    async triggerAlert(alertIndex) {
        return this.sendRequest('trigger_alert', { alert_index: alertIndex });
    }
}

// Usage
const client = new MCPClient();
client.getSessionInfo().then(console.log);
```

## Error Handling

### Standard Error Codes

- `-32700`: Parse error
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

### Error Response Format

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

## Security Considerations

- **Local Access Only**: MCP server binds to localhost
- **No Authentication**: Suitable for local development
- **CORS Enabled**: Cross-origin requests allowed
- **Rate Limiting**: Consider implementing for production use

## Troubleshooting

### Connection Issues

1. **Check if Focus Helper is running**:
   ```bash
   ps aux | grep focushelper
   ```

2. **Verify MCP server is accessible**:
   ```bash
   curl http://localhost:8089/health
   ```

3. **Check port availability**:
   ```bash
   netstat -tulpn | grep :8089
   ```

### Common Errors

- **Connection Refused**: Focus Helper not running or wrong port
- **Method Not Found**: Incorrect method name
- **Invalid Params**: Missing or incorrect parameters
- **Internal Error**: Focus Helper internal issue

---

**Ready to integrate?** Check out our [Examples](/examples/) section for more detailed integration examples and use cases.
