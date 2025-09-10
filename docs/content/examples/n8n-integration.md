---
title: "n8n Integration"
description: "Integrate Focus Helper with n8n workflow automation"
date: 2025-09-09T19:00:00Z
draft: false
weight: 10
---

# n8n Integration with Focus Helper

This guide shows how to integrate Focus Helper with n8n, a powerful workflow automation tool, to create sophisticated focus management workflows.

## Overview

n8n allows you to create complex workflows that can:
- Monitor focus sessions in real-time
- Send notifications to multiple channels
- Trigger actions based on focus patterns
- Integrate with external services
- Create custom dashboards

## Prerequisites

- **n8n installed** (Docker, npm, or self-hosted)
- **Focus Helper running** with MCP server enabled
- **Basic n8n knowledge**

## Setup

### 1. Start Focus Helper with MCP

```bash
focushelper --mcp --docker
```

### 2. Install n8n

#### Using Docker (Recommended)
```bash
docker run -it --rm --name n8n -p 5678:5678 -v ~/.n8n:/home/node/.n8n n8nio/n8n
```

#### Using npm
```bash
npm install n8n -g
n8n start
```

### 3. Access n8n

Open your browser and go to `http://localhost:5678`

## Basic Workflow Examples

### 1. Focus Session Monitor

Create a workflow that monitors focus sessions and sends notifications.

#### Workflow Steps:

1. **HTTP Request Node** - Get session info
2. **IF Node** - Check if session is active
3. **Switch Node** - Route based on session duration
4. **Slack/Email Node** - Send notifications

#### Configuration:

**HTTP Request Node:**
```json
{
  "method": "POST",
  "url": "http://localhost:8089/mcp",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "get_session_info",
    "params": {}
  }
}
```

**IF Node (Check if active):**
```javascript
// Expression
{{ $json.result.is_active === true }}
```

**Switch Node (Duration check):**
```javascript
// Route 1: Short session (< 30 minutes)
{{ $json.result.duration && $json.result.duration.includes('m') && parseInt($json.result.duration) < 30 }}

// Route 2: Medium session (30-60 minutes)
{{ $json.result.duration && ($json.result.duration.includes('h') || parseInt($json.result.duration) >= 30) }}

// Route 3: Long session (> 1 hour)
{{ $json.result.duration && $json.result.duration.includes('h') && parseInt($json.result.duration.split('h')[0]) > 1 }}
```

### 2. Hyperfocus Alert System

Create an alert system that triggers when hyperfocus is detected.

#### Workflow Steps:

1. **Cron Trigger** - Run every 5 minutes
2. **HTTP Request** - Get hyperfocus status
3. **IF Node** - Check if in hyperfocus
4. **HTTP Request** - Trigger alert
5. **Slack Node** - Send notification

#### Configuration:

**Cron Trigger:**
```json
{
  "rule": {
    "interval": [
      {
        "field": "minute",
        "value": "*/5"
      }
    ]
  }
}
```

**HTTP Request (Get hyperfocus status):**
```json
{
  "method": "POST",
  "url": "http://localhost:8089/mcp",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "get_hyperfocus_status",
    "params": {}
  }
}
```

**IF Node (Check hyperfocus):**
```javascript
{{ $json.result.is_hyperfocus === true }}
```

**HTTP Request (Trigger alert):**
```json
{
  "method": "POST",
  "url": "http://localhost:8089/mcp",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "trigger_alert",
    "params": {
      "alert_index": 1
    }
  }
}
```

### 3. Break Reminder System

Create an automated break reminder system.

#### Workflow Steps:

1. **Cron Trigger** - Run every 15 minutes
2. **HTTP Request** - Get session info
3. **Function Node** - Calculate break need
4. **Switch Node** - Route based on break urgency
5. **Multiple Action Nodes** - Send reminders

#### Configuration:

**Function Node (Calculate break need):**
```javascript
const session = $input.first().json.result;
const duration = session.duration;

// Parse duration to minutes
let minutes = 0;
if (duration.includes('h')) {
  const hours = parseInt(duration.split('h')[0]);
  minutes += hours * 60;
}
if (duration.includes('m')) {
  const mins = parseInt(duration.match(/(\d+)m/)?.[1] || 0);
  minutes += mins;
}

// Determine break urgency
let breakUrgency = 'none';
if (minutes > 240) { // 4 hours
  breakUrgency = 'critical';
} else if (minutes > 120) { // 2 hours
  breakUrgency = 'high';
} else if (minutes > 60) { // 1 hour
  breakUrgency = 'medium';
} else if (minutes > 30) { // 30 minutes
  breakUrgency = 'low';
}

return {
  json: {
    ...session,
    durationMinutes: minutes,
    breakUrgency: breakUrgency,
    needsBreak: breakUrgency !== 'none'
  }
};
```

## Advanced Workflow Examples

### 1. Multi-Channel Notification System

Send notifications to multiple channels based on focus patterns.

#### Workflow Steps:

1. **Cron Trigger** - Every 10 minutes
2. **HTTP Request** - Get session info
3. **Function Node** - Analyze focus patterns
4. **Switch Node** - Route to different channels
5. **Parallel Execution** - Send to Slack, Email, Discord

#### Configuration:

**Function Node (Analyze patterns):**
```javascript
const session = $input.first().json.result;
const duration = session.duration;
const hyperfocus = session.hyperfocus_level;

// Parse duration
let minutes = 0;
if (duration.includes('h')) {
  minutes += parseInt(duration.split('h')[0]) * 60;
}
if (duration.includes('m')) {
  minutes += parseInt(duration.match(/(\d+)m/)?.[1] || 0);
}

// Determine notification type
let notificationType = 'info';
let message = `Focus session: ${duration}`;

if (hyperfocus) {
  notificationType = 'warning';
  message = `⚠️ Hyperfocus detected: ${hyperfocus} level, ${duration} duration`;
}

if (minutes > 180) { // 3 hours
  notificationType = 'critical';
  message = `🚨 Critical: Focus session too long (${duration})`;
}

return {
  json: {
    notificationType,
    message,
    duration,
    minutes,
    hyperfocus,
    timestamp: new Date().toISOString()
  }
};
```

### 2. Focus Analytics Dashboard

Create a dashboard that tracks focus patterns over time.

#### Workflow Steps:

1. **Cron Trigger** - Every hour
2. **HTTP Request** - Get session info
3. **Function Node** - Process data
4. **Google Sheets Node** - Store data
5. **Webhook Node** - Update dashboard

#### Configuration:

**Function Node (Process analytics):**
```javascript
const session = $input.first().json.result;
const now = new Date();

// Calculate session metrics
const duration = session.duration;
const isActive = session.is_active;
const hyperfocus = session.hyperfocus_level;

// Parse duration to minutes
let minutes = 0;
if (duration.includes('h')) {
  minutes += parseInt(duration.split('h')[0]) * 60;
}
if (duration.includes('m')) {
  minutes += parseInt(duration.match(/(\d+)m/)?.[1] || 0);
}

// Calculate productivity score
let productivityScore = 0;
if (isActive) {
  productivityScore = Math.min(minutes / 60, 8); // Max 8 hours
}

// Determine focus quality
let focusQuality = 'normal';
if (hyperfocus === 'critical') {
  focusQuality = 'unhealthy';
} else if (hyperfocus === 'high') {
  focusQuality = 'intense';
} else if (hyperfocus === 'medium') {
  focusQuality = 'good';
} else if (hyperfocus === 'low') {
  focusQuality = 'light';
}

return {
  json: {
    timestamp: now.toISOString(),
    date: now.toISOString().split('T')[0],
    hour: now.getHours(),
    duration: duration,
    durationMinutes: minutes,
    isActive: isActive,
    hyperfocus: hyperfocus,
    focusQuality: focusQuality,
    productivityScore: productivityScore,
    subject: session.subject || 'Unknown'
  }
};
```

### 3. Smart Break Scheduler

Create an intelligent break scheduler that adapts to focus patterns.

#### Workflow Steps:

1. **Cron Trigger** - Every 5 minutes
2. **HTTP Request** - Get session info
3. **Function Node** - Calculate break timing
4. **Switch Node** - Route based on break need
5. **HTTP Request** - Trigger appropriate alert
6. **Calendar Node** - Schedule break time

#### Configuration:

**Function Node (Smart break calculation):**
```javascript
const session = $input.first().json.result;
const duration = session.duration;
const hyperfocus = session.hyperfocus_level;

// Parse duration
let minutes = 0;
if (duration.includes('h')) {
  minutes += parseInt(duration.split('h')[0]) * 60;
}
if (duration.includes('m')) {
  minutes += parseInt(duration.match(/(\d+)m/)?.[1] || 0);
}

// Calculate break recommendations
let breakRecommendation = null;
let alertLevel = -1;

if (minutes > 240) { // 4+ hours
  breakRecommendation = {
    type: 'emergency',
    duration: 30,
    message: 'Emergency break needed - 30 minutes recommended'
  };
  alertLevel = 3; // Critical
} else if (minutes > 120) { // 2+ hours
  breakRecommendation = {
    type: 'extended',
    duration: 20,
    message: 'Extended break recommended - 20 minutes'
  };
  alertLevel = 2; // High
} else if (minutes > 60) { // 1+ hour
  breakRecommendation = {
    type: 'standard',
    duration: 15,
    message: 'Standard break recommended - 15 minutes'
  };
  alertLevel = 1; // Medium
} else if (minutes > 30) { // 30+ minutes
  breakRecommendation = {
    type: 'micro',
    duration: 5,
    message: 'Micro break recommended - 5 minutes'
  };
  alertLevel = 0; // Low
}

// Adjust based on hyperfocus level
if (hyperfocus === 'critical') {
  breakRecommendation.duration = Math.max(breakRecommendation.duration, 30);
  breakRecommendation.message += ' (Critical hyperfocus detected)';
} else if (hyperfocus === 'high') {
  breakRecommendation.duration = Math.max(breakRecommendation.duration, 20);
  breakRecommendation.message += ' (High hyperfocus detected)';
}

return {
  json: {
    ...session,
    durationMinutes: minutes,
    breakRecommendation,
    alertLevel,
    shouldTriggerAlert: alertLevel >= 0
  }
};
```

## Integration with External Services

### 1. Slack Integration

Send focus updates to Slack channels.

**Slack Node Configuration:**
```json
{
  "channel": "#focus-updates",
  "text": "{{ $json.message }}",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Focus Session Update*\n*Duration:* {{ $json.duration }}\n*Status:* {{ $json.is_active ? 'Active' : 'Inactive' }}\n*Hyperfocus:* {{ $json.hyperfocus_level || 'None' }}"
      }
    }
  ]
}
```

### 2. Email Notifications

Send detailed focus reports via email.

**Email Node Configuration:**
```json
{
  "to": "user@example.com",
  "subject": "Focus Session Report - {{ $json.timestamp }}",
  "html": "<h2>Focus Session Report</h2><p><strong>Duration:</strong> {{ $json.duration }}</p><p><strong>Status:</strong> {{ $json.is_active ? 'Active' : 'Inactive' }}</p><p><strong>Hyperfocus Level:</strong> {{ $json.hyperfocus_level || 'None' }}</p><p><strong>Subject:</strong> {{ $json.subject }}</p>"
}
```

### 3. Google Calendar Integration

Schedule breaks and focus sessions in Google Calendar.

**Google Calendar Node Configuration:**
```json
{
  "calendarId": "primary",
  "summary": "Focus Break - {{ $json.breakRecommendation.type }}",
  "description": "{{ $json.breakRecommendation.message }}",
  "start": "{{ $now }}",
  "end": "{{ $now.plus({ minutes: $json.breakRecommendation.duration }) }}"
}
```

### 4. Discord Integration

Send focus updates to Discord channels.

**Discord Node Configuration:**
```json
{
  "channel": "focus-updates",
  "content": "{{ $json.message }}",
  "embeds": [
    {
      "title": "Focus Session Update",
      "color": 3447003,
      "fields": [
        {
          "name": "Duration",
          "value": "{{ $json.duration }}",
          "inline": true
        },
        {
          "name": "Status",
          "value": "{{ $json.is_active ? 'Active' : 'Inactive' }}",
          "inline": true
        },
        {
          "name": "Hyperfocus",
          "value": "{{ $json.hyperfocus_level || 'None' }}",
          "inline": true
        }
      ],
      "timestamp": "{{ $json.timestamp }}"
    }
  ]
}
```

## Error Handling and Monitoring

### 1. Error Handling Workflow

Create a workflow that handles errors and sends alerts.

#### Workflow Steps:

1. **Error Trigger** - Catch all errors
2. **Function Node** - Process error information
3. **Slack Node** - Send error notification
4. **HTTP Request** - Log error to external service

#### Configuration:

**Error Trigger:**
```json
{
  "errorWorkflow": true
}
```

**Function Node (Process error):**
```javascript
const error = $input.first().json;

return {
  json: {
    error: error.message,
    node: error.node,
    timestamp: new Date().toISOString(),
    workflow: error.workflow,
    severity: 'high'
  }
};
```

### 2. Health Monitoring

Monitor Focus Helper and n8n health.

#### Workflow Steps:

1. **Cron Trigger** - Every minute
2. **HTTP Request** - Check Focus Helper health
3. **HTTP Request** - Check n8n health
4. **Function Node** - Analyze health status
5. **Switch Node** - Route based on health
6. **Notification Node** - Send alerts if unhealthy

#### Configuration:

**HTTP Request (Focus Helper health):**
```json
{
  "method": "GET",
  "url": "http://localhost:8088/health"
}
```

**HTTP Request (n8n health):**
```json
{
  "method": "GET",
  "url": "http://localhost:5678/healthz"
}
```

## Best Practices

### 1. Workflow Organization

- **Use descriptive names** for workflows and nodes
- **Add comments** to complex function nodes
- **Group related workflows** in folders
- **Use consistent naming conventions**

### 2. Error Handling

- **Always include error handling** in workflows
- **Use try-catch blocks** in function nodes
- **Set up error notifications** for critical workflows
- **Test error scenarios** regularly

### 3. Performance

- **Use appropriate trigger intervals** (not too frequent)
- **Cache data** when possible
- **Use parallel execution** for independent operations
- **Monitor workflow execution times**

### 4. Security

- **Use environment variables** for sensitive data
- **Validate input data** in function nodes
- **Use HTTPS** for external API calls
- **Implement rate limiting** for external services

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if Focus Helper is running
   - Verify MCP server port (8089)
   - Check firewall settings

2. **Authentication Errors**
   - Verify API keys and tokens
   - Check service permissions
   - Update expired credentials

3. **Workflow Execution Errors**
   - Check node configurations
   - Verify data formats
   - Review error logs

4. **Performance Issues**
   - Reduce trigger frequency
   - Optimize function node code
   - Check resource usage

### Debug Tips

1. **Use Debug Mode** in n8n
2. **Add logging** in function nodes
3. **Test individual nodes** before full workflow
4. **Monitor execution logs**
5. **Use webhook nodes** for testing

---

**Ready to automate your focus management?** Start with the basic workflows and gradually add complexity as you become more comfortable with n8n and Focus Helper integration.
