# Cursor Integration with Focus Helper

This guide explains how to connect Cursor with your Focus Helper application using the Model Context Protocol (MCP) server.

## Overview

The integration allows Cursor to:
- Monitor your current focus session
- Check hyperfocus status and duration
- View available alert levels
- Trigger specific alert levels
- Get real-time session information

## Prerequisites

1. **Focus Helper running with MCP server**
2. **Node.js installed** (for the MCP bridge)
3. **Cursor with MCP support**

## Setup Instructions

### 1. Start Focus Helper with MCP Server

```bash
# Start focus-helper with MCP server enabled
focushelper --mcp --docker

# Or with custom port
focushelper --mcp --mcp-port 8089
```

Verify the MCP server is running:
```bash
curl http://localhost:8089/health
```

### 2. Configure Cursor MCP Settings

1. **Open Cursor Settings** (Cmd/Ctrl + ,)
2. **Search for "MCP"** in settings
3. **Add the following configuration**:

```json
{
  "mcpServers": {
    "focus-helper": {
      "command": "node",
      "args": ["/home/robertocpaes/Projects/Pessoal/focus-helper/mcp-server.js"],
      "env": {}
    }
  }
}
```

**Alternative method**: Create a `.cursorrules` file in your project root:

```json
{
  "mcpServers": {
    "focus-helper": {
      "command": "node",
      "args": ["/home/robertocpaes/Projects/Pessoal/focus-helper/mcp-server.js"],
      "env": {}
    }
  }
}
```

### 3. Restart Cursor

After adding the MCP configuration, restart Cursor to load the new server.

## Available Tools in Cursor

Once connected, you'll have access to these tools in Cursor:

### 1. `get_session_info`
Get current focus session information including:
- Session ID and subject
- Start time and current time
- Session duration
- Activity status
- Hyperfocus information
- Idle duration

**Usage in Cursor**: Ask "What's my current focus session status?"

### 2. `get_alert_levels`
Get available alert levels and their thresholds:
- Alert level names (low, medium, high, critical)
- Threshold durations
- Enabled status
- Tolerance settings

**Usage in Cursor**: Ask "What alert levels are available?"

### 3. `trigger_alert`
Trigger a specific alert level by index:
- Low (index 0): 30 minutes
- Medium (index 1): 1.5 hours
- High (index 2): 3 hours
- Critical (index 3): 6 hours

**Usage in Cursor**: Ask "Trigger a low alert" or "Send me a break reminder"

### 4. `get_hyperfocus_status`
Get current hyperfocus status:
- Whether you're in hyperfocus
- Current hyperfocus level
- Hyperfocus start time
- Hyperfocus duration

**Usage in Cursor**: Ask "Am I in hyperfocus right now?"

### 5. `ping`
Test connection to the focus helper MCP server.

**Usage in Cursor**: Ask "Is the focus helper connected?"

## Example Interactions

### Check Focus Status
```
You: "What's my current focus session status?"
Cursor: [Uses get_session_info tool]
Response: "You've been focused for 45 minutes on 'General Use'. 
          You're currently in a 'low' hyperfocus state that started 15 minutes ago. 
          Your last activity was 2 minutes ago."
```

### Request Break Reminder
```
You: "I need a break reminder"
Cursor: [Uses trigger_alert tool with index 0]
Response: "I've triggered a low-level alert for you. 
          This will play a sound and provide a break reminder."
```

### Check Available Alerts
```
You: "What alert levels can you trigger?"
Cursor: [Uses get_alert_levels tool]
Response: "Available alert levels:
          - Low (30 minutes): Basic break reminder
          - Medium (1.5 hours): More urgent reminder
          - High (3 hours): Strong intervention needed
          - Critical (6 hours): Emergency intervention"
```

### Monitor Hyperfocus
```
You: "Am I in hyperfocus right now?"
Cursor: [Uses get_hyperfocus_status tool]
Response: "Yes, you're currently in a 'medium' hyperfocus state that started 2 hours ago. 
          This is getting quite intense - you might want to consider taking a break."
```

## Troubleshooting

### MCP Server Not Connecting

1. **Check if Focus Helper is running**:
   ```bash
   ps aux | grep focushelper
   ```

2. **Verify MCP server is accessible**:
   ```bash
   curl http://localhost:8089/health
   ```

3. **Check Cursor MCP logs**:
   - Open Cursor Developer Tools (Help > Toggle Developer Tools)
   - Look for MCP-related errors in the console

### Node.js Bridge Issues

1. **Test the bridge manually**:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}' | node mcp-server.js
   ```

2. **Check Node.js version**:
   ```bash
   node --version
   ```

### Focus Helper Not Responding

1. **Restart Focus Helper**:
   ```bash
   pkill -f focushelper
   focushelper --mcp --docker
   ```

2. **Check logs**:
   ```bash
   tail -f focus_helper.log
   ```

## Advanced Configuration

### Custom Port Configuration

If you want to use a different port for the MCP server:

1. **Start Focus Helper with custom port**:
   ```bash
   focushelper --mcp --mcp-port 9090
   ```

2. **Update the bridge server** (edit `mcp-server.js`):
   ```javascript
   this.baseUrl = 'http://localhost:9090';
   ```

### Multiple Focus Helper Instances

To connect to multiple Focus Helper instances:

```json
{
  "mcpServers": {
    "focus-helper-main": {
      "command": "node",
      "args": ["/path/to/mcp-server.js"],
      "env": { "FOCUS_HELPER_PORT": "8089" }
    },
    "focus-helper-work": {
      "command": "node", 
      "args": ["/path/to/mcp-server.js"],
      "env": { "FOCUS_HELPER_PORT": "8090" }
    }
  }
}
```

## Security Considerations

- The MCP server runs on localhost only
- No authentication is required (suitable for local use)
- For production use, consider adding authentication
- The bridge server has read-only access to session data
- Alert triggering requires explicit user confirmation

## Benefits

1. **Context-Aware Assistance**: Cursor knows your focus state
2. **Proactive Reminders**: Can suggest breaks based on session duration
3. **Hyperfocus Monitoring**: Helps prevent unhealthy hyperfocus
4. **Seamless Integration**: Works directly within your coding environment
5. **Real-time Updates**: Always has current session information

## Example Workflows

### Morning Focus Session
```
1. Start focus-helper: focushelper --mcp
2. Open Cursor
3. Ask: "What's my focus status?"
4. Work on code
5. Ask: "How long have I been focused?"
6. When ready: "Trigger a break reminder"
```

### Hyperfocus Management
```
1. Monitor: "Am I in hyperfocus?"
2. Check duration: "How long have I been hyperfocused?"
3. If too long: "Trigger a high alert"
4. Take break
5. Resume: "What's my current status?"
```

This integration makes Cursor a powerful tool for managing your focus and productivity while coding!
