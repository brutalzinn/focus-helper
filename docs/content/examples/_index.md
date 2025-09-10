---
title: "Examples"
description: "Practical examples and use cases"
date: 2025-09-09T19:00:00Z
draft: false
weight: 50
---

# Examples

This section provides practical examples and use cases for Focus Helper, demonstrating how to integrate it with various tools and workflows.

## MCP Client Examples

### Go Client

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    client := NewMCPClient("http://localhost:8089")
    
    // Get session information
    session, err := client.GetSessionInfo()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Session duration: %s\n", session.Duration)
    fmt.Printf("Is active: %t\n", session.IsActive)
    
    // Check hyperfocus status
    if session.HyperfocusLevel != "" {
        fmt.Printf("Hyperfocus level: %s\n", session.HyperfocusLevel)
        fmt.Printf("Hyperfocus duration: %s\n", session.HyperfocusDuration)
    }
    
    // Trigger alert if session is too long
    if session.Duration > 2*time.Hour {
        fmt.Println("Session too long, triggering alert...")
        if err := client.TriggerAlert(1); err != nil {
            panic(err)
        }
    }
}
```

### Python Client

```python
import requests
import json
import time
from datetime import datetime, timedelta

class FocusHelperClient:
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
    
    def trigger_alert(self, alert_index):
        payload = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "trigger_alert",
            "params": {"alert_index": alert_index}
        }
        
        response = self.session.post(f"{self.base_url}/mcp", json=payload)
        return response.json()

# Usage example
client = FocusHelperClient()

# Monitor session every 5 minutes
while True:
    session = client.get_session_info()
    if session.get('result'):
        duration = session['result']['duration']
        print(f"Current session duration: {duration}")
        
        # Trigger alert if session is too long
        if duration and 'h' in duration and int(duration.split('h')[0]) > 2:
            print("Session too long, triggering alert...")
            client.trigger_alert(1)
    
    time.sleep(300)  # Wait 5 minutes
```

### JavaScript/Node.js Client

```javascript
const http = require('http');

class FocusHelperClient {
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

// Usage example
const client = new FocusHelperClient();

async function monitorSession() {
    try {
        const session = await client.getSessionInfo();
        console.log('Session info:', session.result);
        
        // Trigger alert if needed
        if (session.result.duration > '2h') {
            console.log('Triggering alert...');
            await client.triggerAlert(1);
        }
    } catch (error) {
        console.error('Error:', error);
    }
}

// Monitor every 5 minutes
setInterval(monitorSession, 5 * 60 * 1000);
```

## Cursor IDE Integration

### MCP Server Configuration

Create `~/.cursor/mcp-servers.json`:

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

### Example Interactions

**Check Focus Status:**
```
You: "What's my current focus session status?"
Cursor: [Uses get_session_info tool]
Response: "You've been focused for 45 minutes on 'General Use'. 
          You're currently in a 'low' hyperfocus state that started 15 minutes ago."
```

**Request Break Reminder:**
```
You: "I need a break reminder"
Cursor: [Uses trigger_alert tool with index 0]
Response: "I've triggered a low-level alert for you. 
          This will play a sound and provide a break reminder."
```

**Monitor Hyperfocus:**
```
You: "Am I in hyperfocus right now?"
Cursor: [Uses get_hyperfocus_status tool]
Response: "Yes, you're currently in a 'medium' hyperfocus state that started 2 hours ago. 
          This is getting quite intense - you might want to consider taking a break."
```

## Automation Scripts

### Break Reminder Script

```bash
#!/bin/bash
# Auto-trigger break reminder after 1 hour

while true; do
    DURATION=$(curl -s -X POST http://localhost:8089/mcp \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"get_session_info","params":{}}' \
        | jq -r '.result.duration' | sed 's/m.*//')
    
    if [ "$DURATION" -gt 60 ]; then
        echo "Session too long, triggering break reminder..."
        curl -s -X POST http://localhost:8089/mcp \
            -H "Content-Type: application/json" \
            -d '{"jsonrpc":"2.0","id":2,"method":"trigger_alert","params":{"alert_index":0}}'
        break
    fi
    
    sleep 300  # Check every 5 minutes
done
```

### Session Logger

```python
import requests
import json
import time
from datetime import datetime

def log_session_data():
    client = requests.Session()
    client.headers.update({"Content-Type": "application/json"})
    
    while True:
        try:
            # Get session info
            response = client.post('http://localhost:8089/mcp', json={
                "jsonrpc": "2.0",
                "id": 1,
                "method": "get_session_info",
                "params": {}
            })
            
            session_data = response.json()
            
            if session_data.get('result'):
                # Log to file
                with open('session_log.txt', 'a') as f:
                    f.write(f"{datetime.now()}: {json.dumps(session_data['result'])}\n")
                
                print(f"Logged session data: {session_data['result']['duration']}")
            
        except Exception as e:
            print(f"Error: {e}")
        
        time.sleep(60)  # Log every minute

if __name__ == "__main__":
    log_session_data()
```

### Hyperfocus Monitor

```javascript
const http = require('http');
const fs = require('fs');

class HyperfocusMonitor {
    constructor() {
        this.alertThresholds = {
            'low': 30,      // 30 minutes
            'medium': 60,   // 1 hour
            'high': 120,    // 2 hours
            'critical': 240 // 4 hours
        };
    }
    
    async checkHyperfocus() {
        try {
            const session = await this.getSessionInfo();
            const hyperfocus = await this.getHyperfocusStatus();
            
            if (hyperfocus.result.is_hyperfocus) {
                const duration = this.parseDuration(hyperfocus.result.duration);
                const level = hyperfocus.result.level;
                
                console.log(`Hyperfocus detected: ${level} level, ${duration} minutes`);
                
                // Check if threshold exceeded
                if (duration > this.alertThresholds[level]) {
                    console.log(`Threshold exceeded for ${level} level!`);
                    await this.triggerAlert(this.getAlertIndex(level));
                }
            }
        } catch (error) {
            console.error('Error checking hyperfocus:', error);
        }
    }
    
    async getSessionInfo() {
        return this.sendRequest('get_session_info');
    }
    
    async getHyperfocusStatus() {
        return this.sendRequest('get_hyperfocus_status');
    }
    
    async triggerAlert(alertIndex) {
        return this.sendRequest('trigger_alert', { alert_index: alertIndex });
    }
    
    sendRequest(method, params = {}) {
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
    
    parseDuration(duration) {
        // Parse duration string like "1h30m" to minutes
        const hours = (duration.match(/(\d+)h/) || [0, 0])[1];
        const minutes = (duration.match(/(\d+)m/) || [0, 0])[1];
        return parseInt(hours) * 60 + parseInt(minutes);
    }
    
    getAlertIndex(level) {
        const levels = ['low', 'medium', 'high', 'critical'];
        return levels.indexOf(level);
    }
}

// Start monitoring
const monitor = new HyperfocusMonitor();
setInterval(() => monitor.checkHyperfocus(), 60000); // Check every minute
```

## Docker Compose Examples

### Development Setup

```yaml
version: '3.8'

services:
  focus-helper:
    build: .
    container_name: focus-helper-dev
    ports:
      - "8088:8088"
      - "8089:8089"
    environment:
      - FOCUSHELPER_DOCKER_MODE=true
      - FOCUSHELPER_DEBUG=true
    volumes:
      - ./profiles.json:/home/focushelper/.config/focushelper/profiles.json
      - focus-helper-data:/home/focushelper/.config/focushelper
    devices:
      - "/dev/snd:/dev/snd"
    privileged: true
    group_add:
      - audio

volumes:
  focus-helper-data:
```

### Production Setup

```yaml
version: '3.8'

services:
  focus-helper:
    image: focus-helper:latest
    container_name: focus-helper-prod
    restart: unless-stopped
    ports:
      - "8088:8088"
      - "8089:8089"
    environment:
      - FOCUSHELPER_DOCKER_MODE=true
      - FOCUSHELPER_MCP_SERVER_ENABLED=true
    volumes:
      - focus-helper-data:/home/focushelper/.config/focushelper
    devices:
      - "/dev/snd:/dev/snd"
    privileged: true
    group_add:
      - audio
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8088/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  focus-helper-data:
    driver: local
```

## Webhook Integration

### Slack Notifications

```python
import requests
import json
import time

class FocusHelperSlackBot:
    def __init__(self, webhook_url):
        self.webhook_url = webhook_url
        self.focus_helper_url = "http://localhost:8089"
    
    def send_slack_message(self, message):
        payload = {"text": message}
        requests.post(self.webhook_url, json=payload)
    
    def monitor_focus_sessions(self):
        while True:
            try:
                # Get session info
                response = requests.post(f"{self.focus_helper_url}/mcp", json={
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "get_session_info",
                    "params": {}
                })
                
                session = response.json()
                
                if session.get('result'):
                    duration = session['result']['duration']
                    is_active = session['result']['is_active']
                    
                    if is_active and duration:
                        # Send notification if session is long
                        if 'h' in duration and int(duration.split('h')[0]) > 2:
                            self.send_slack_message(
                                f"⚠️ Long focus session detected: {duration}"
                            )
                
            except Exception as e:
                print(f"Error: {e}")
            
            time.sleep(300)  # Check every 5 minutes

# Usage
bot = FocusHelperSlackBot("https://hooks.slack.com/services/YOUR/WEBHOOK/URL")
bot.monitor_focus_sessions()
```

### Discord Integration

```javascript
const http = require('http');
const { Webhook } = require('discord-webhook-node');

class FocusHelperDiscordBot {
    constructor(webhookUrl) {
        this.webhook = new Webhook(webhookUrl);
        this.focusHelperUrl = 'http://localhost:8089';
    }
    
    async sendDiscordMessage(message) {
        await this.webhook.send(message);
    }
    
    async monitorSessions() {
        while (true) {
            try {
                const session = await this.getSessionInfo();
                
                if (session.result && session.result.is_active) {
                    const duration = session.result.duration;
                    const hyperfocus = session.result.hyperfocus_level;
                    
                    if (hyperfocus) {
                        await this.sendDiscordMessage(
                            `🧠 Hyperfocus detected: ${hyperfocus} level, ${duration} duration`
                        );
                    }
                }
            } catch (error) {
                console.error('Error:', error);
            }
            
            await new Promise(resolve => setTimeout(resolve, 60000)); // Wait 1 minute
        }
    }
    
    async getSessionInfo() {
        return this.sendRequest('get_session_info');
    }
    
    sendRequest(method, params = {}) {
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
}

// Usage
const bot = new FocusHelperDiscordBot('https://discord.com/api/webhooks/YOUR/WEBHOOK/URL');
bot.monitorSessions();
```

## Research and Analytics

### Session Data Export

```python
import requests
import json
import csv
from datetime import datetime, timedelta

class FocusHelperAnalytics:
    def __init__(self):
        self.base_url = "http://localhost:8089"
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})
    
    def export_session_data(self, days=7):
        """Export session data for the last N days"""
        data = []
        
        for i in range(days):
            date = datetime.now() - timedelta(days=i)
            # In a real implementation, you'd query historical data
            # For now, we'll just get current session info
            session_info = self.get_session_info()
            if session_info.get('result'):
                data.append({
                    'date': date.strftime('%Y-%m-%d'),
                    'duration': session_info['result']['duration'],
                    'is_active': session_info['result']['is_active'],
                    'hyperfocus_level': session_info['result'].get('hyperfocus_level', ''),
                    'subject': session_info['result']['subject']
                })
        
        # Export to CSV
        with open('focus_sessions.csv', 'w', newline='') as csvfile:
            fieldnames = ['date', 'duration', 'is_active', 'hyperfocus_level', 'subject']
            writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
            writer.writeheader()
            writer.writerows(data)
        
        print(f"Exported {len(data)} session records to focus_sessions.csv")
    
    def get_session_info(self):
        response = self.session.post(f"{self.base_url}/mcp", json={
            "jsonrpc": "2.0",
            "id": 1,
            "method": "get_session_info",
            "params": {}
        })
        return response.json()

# Usage
analytics = FocusHelperAnalytics()
analytics.export_session_data(days=30)
```

---

**Need more examples?** Check out our [GitHub repository](https://github.com/robertocpaes/focus-helper) for additional examples and community contributions.
