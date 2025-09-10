---
title: "n8n Integration Guide"
description: "Complete guide to integrating Focus Helper with n8n for emergency notifications and data collection"
date: 2024-09-09
weight: 15
---

# n8n Integration Guide

This guide shows you how to integrate Focus Helper with n8n to create powerful automation workflows for emergency notifications, data collection, and family communication.

## 🚨 Emergency Notification Workflows

### Level 1: Gentle Reminder
**Trigger**: Hyperfocus detected for 2+ hours
**Actions**:
- Voice reminder through Focus Helper
- Log event in database
- Send gentle notification to user's phone

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-alert",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Check Alert Level",
      "type": "n8n-nodes-base.if",
      "parameters": {
        "conditions": {
          "string": [
            {
              "value1": "={{ $json.level }}",
              "operation": "equal",
              "value2": "1"
            }
          ]
        }
      }
    },
    {
      "name": "Send Gentle Reminder",
      "type": "n8n-nodes-base.telegram",
      "parameters": {
        "chatId": "{{ $json.user_telegram_id }}",
        "text": "⏰ Gentle reminder: You've been focused for {{ $json.duration }}. Time for a break?"
      }
    }
  ]
}
```

### Level 2: Escalated Alert
**Trigger**: No response to Level 1 for 30 minutes
**Actions**:
- Send SMS to user
- Email notification with details
- Log in family dashboard

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-alert",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Check Alert Level",
      "type": "n8n-nodes-base.if",
      "parameters": {
        "conditions": {
          "string": [
            {
              "value1": "={{ $json.level }}",
              "operation": "equal",
              "value2": "2"
            }
          ]
        }
      }
    },
    {
      "name": "Send SMS",
      "type": "n8n-nodes-base.twilio",
      "parameters": {
        "to": "{{ $json.user_phone }}",
        "message": "🚨 Focus Helper Alert: You've been focused for {{ $json.duration }}. Please take a break and check in."
      }
    },
    {
      "name": "Send Email",
      "type": "n8n-nodes-base.gmail",
      "parameters": {
        "to": "{{ $json.user_email }}",
        "subject": "Focus Helper Alert - Level 2",
        "message": "You've been in hyperfocus for {{ $json.duration }}. Please take a break and respond to this message."
      }
    }
  ]
}
```

### Level 3: Family Notification
**Trigger**: No response to Level 2 for 1 hour
**Actions**:
- Notify family members via SMS
- Send email to emergency contacts
- Update family dashboard

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-alert",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Check Alert Level",
      "type": "n8n-nodes-base.if",
      "parameters": {
        "conditions": {
          "string": [
            {
              "value1": "={{ $json.level }}",
              "operation": "equal",
              "value2": "3"
            }
          ]
        }
      }
    },
    {
      "name": "Notify Family",
      "type": "n8n-nodes-base.splitInBatches",
      "parameters": {
        "batchSize": 1,
        "options": {}
      }
    },
    {
      "name": "Send Family SMS",
      "type": "n8n-nodes-base.twilio",
      "parameters": {
        "to": "={{ $json.family_phone }}",
        "message": "👨‍👩‍👧‍👦 Family Alert: {{ $json.user_name }} has been in hyperfocus for {{ $json.duration }}. Please check on them."
      }
    },
    {
      "name": "Update Dashboard",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "https://your-dashboard.com/api/alert",
        "method": "POST",
        "body": {
          "user_id": "{{ $json.user_id }}",
          "alert_level": "3",
          "duration": "{{ $json.duration }}",
          "timestamp": "{{ $json.timestamp }}"
        }
      }
    }
  ]
}
```

### Level 4: Emergency Protocol
**Trigger**: No response to Level 3 for 2 hours
**Actions**:
- Call emergency contacts
- Send urgent notifications
- Contact healthcare provider if configured
- Log critical event

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-alert",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Check Alert Level",
      "type": "n8n-nodes-base.if",
      "parameters": {
        "conditions": {
          "string": [
            {
              "value1": "={{ $json.level }}",
              "operation": "equal",
              "value2": "4"
            }
          ]
        }
      }
    },
    {
      "name": "Call Emergency Contact",
      "type": "n8n-nodes-base.twilio",
      "parameters": {
        "to": "{{ $json.emergency_phone }}",
        "message": "🚨 EMERGENCY: {{ $json.user_name }} has been in hyperfocus for {{ $json.duration }} and hasn't responded to alerts. Please check on them immediately."
      }
    },
    {
      "name": "Notify Healthcare Provider",
      "type": "n8n-nodes-base.gmail",
      "parameters": {
        "to": "{{ $json.doctor_email }}",
        "subject": "URGENT: Patient Alert - Focus Helper",
        "message": "Patient {{ $json.user_name }} has been in hyperfocus for {{ $json.duration }} and hasn't responded to safety alerts. Please check on them."
      }
    },
    {
      "name": "Log Critical Event",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "https://your-logging-service.com/api/critical",
        "method": "POST",
        "body": {
          "event_type": "emergency_alert",
          "user_id": "{{ $json.user_id }}",
          "duration": "{{ $json.duration }}",
          "timestamp": "{{ $json.timestamp }}",
          "alert_level": "4"
        }
      }
    }
  ]
}
```

## 📊 Data Collection & Analytics

### Daily Activity Tracking
**Purpose**: Collect data about focus patterns and health metrics

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-data",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Store in Database",
      "type": "n8n-nodes-base.postgres",
      "parameters": {
        "operation": "insert",
        "table": "focus_sessions",
        "columns": "user_id, start_time, end_time, duration, activity_type, health_metrics",
        "values": "={{ $json.user_id }}, {{ $json.start_time }}, {{ $json.end_time }}, {{ $json.duration }}, {{ $json.activity_type }}, {{ $json.health_metrics }}"
      }
    },
    {
      "name": "Generate Daily Report",
      "type": "n8n-nodes-base.cron",
      "parameters": {
        "rule": "0 20 * * *"
      }
    },
    {
      "name": "Send Daily Summary",
      "type": "n8n-nodes-base.gmail",
      "parameters": {
        "to": "{{ $json.user_email }}",
        "subject": "Daily Focus Report - {{ $json.date }}",
        "message": "Today's focus summary: {{ $json.summary }}"
      }
    }
  ]
}
```

### Weekly Family Reports
**Purpose**: Keep family informed about progress and patterns

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Weekly Trigger",
      "type": "n8n-nodes-base.cron",
      "parameters": {
        "rule": "0 9 * * 1"
      }
    },
    {
      "name": "Generate Weekly Report",
      "type": "n8n-nodes-base.function",
      "parameters": {
        "functionCode": "// Generate weekly report from database data\nconst report = {\n  total_focus_time: $json.total_focus_time,\n  average_session_length: $json.avg_session_length,\n  break_frequency: $json.break_frequency,\n  health_metrics: $json.health_metrics,\n  improvements: $json.improvements\n};\n\nreturn { report };"
      }
    },
    {
      "name": "Send to Family",
      "type": "n8n-nodes-base.gmail",
      "parameters": {
        "to": "{{ $json.family_email }}",
        "subject": "Weekly Focus Report - {{ $json.user_name }}",
        "message": "{{ $json.report }}"
      }
    }
  ]
}
```

## 🏠 Smart Home Integration

### Focus Mode Automation
**Purpose**: Adjust home environment based on focus levels

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-hyperfocus",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Adjust Lighting",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "http://homeassistant.local:8123/api/services/light/turn_on",
        "method": "POST",
        "headers": {
          "Authorization": "Bearer {{ $json.ha_token }}",
          "Content-Type": "application/json"
        },
        "body": {
          "entity_id": "light.workspace",
          "brightness": 80,
          "color_temp": 4000
        }
      }
    },
    {
      "name": "Set Temperature",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "http://homeassistant.local:8123/api/services/climate/set_temperature",
        "method": "POST",
        "headers": {
          "Authorization": "Bearer {{ $json.ha_token }}",
          "Content-Type": "application/json"
        },
        "body": {
          "entity_id": "climate.workspace",
          "temperature": 22
        }
      }
    },
    {
      "name": "Enable Focus Mode",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "http://homeassistant.local:8123/api/services/input_boolean/turn_on",
        "method": "POST",
        "headers": {
          "Authorization": "Bearer {{ $json.ha_token }}",
          "Content-Type": "application/json"
        },
        "body": {
          "entity_id": "input_boolean.focus_mode"
        }
      }
    }
  ]
}
```

## 📱 Mobile App Integration

### Push Notifications
**Purpose**: Send real-time notifications to mobile devices

**n8n Workflow**:
```json
{
  "nodes": [
    {
      "name": "Webhook Trigger",
      "type": "n8n-nodes-base.webhook",
      "parameters": {
        "path": "focus-helper-notification",
        "httpMethod": "POST"
      }
    },
    {
      "name": "Send Push Notification",
      "type": "n8n-nodes-base.httpRequest",
      "parameters": {
        "url": "https://fcm.googleapis.com/fcm/send",
        "method": "POST",
        "headers": {
          "Authorization": "key={{ $json.fcm_server_key }}",
          "Content-Type": "application/json"
        },
        "body": {
          "to": "{{ $json.device_token }}",
          "notification": {
            "title": "{{ $json.title }}",
            "body": "{{ $json.message }}",
            "icon": "focus_helper_icon"
          },
          "data": {
            "alert_level": "{{ $json.alert_level }}",
            "duration": "{{ $json.duration }}"
          }
        }
      }
    }
  ]
}
```

## 🔧 Configuration Examples

### Focus Helper Configuration
```json
{
  "name": "n8n_integration",
  "webhook": {
    "enabled": true,
    "url": "http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0",
    "events": [
      "hyperfocus_detected",
      "alert_triggered",
      "session_started",
      "session_ended",
      "emergency_activated"
    ]
  },
  "alerts": {
    "levels": [
      {
        "level": 1,
        "duration": "2h",
        "actions": ["voice_reminder", "log_event"]
      },
      {
        "level": 2,
        "duration": "4h",
        "actions": ["voice_reminder", "sms_user", "email_user"]
      },
      {
        "level": 3,
        "duration": "6h",
        "actions": ["notify_family", "update_dashboard"]
      },
      {
        "level": 4,
        "duration": "8h",
        "actions": ["call_emergency", "notify_healthcare"]
      }
    ]
  }
}
```

### n8n Environment Variables
```bash
# Focus Helper Integration
FOCUSHELPER_WEBHOOK_URL=http://192.168.0.47:5678/webhook-test/4b89be2a-29bf-4fd8-b716-11171b8c60f0
FOCUSHELPER_API_KEY=your_api_key_here

# Notification Services
TWILIO_ACCOUNT_SID=your_twilio_sid
TWILIO_AUTH_TOKEN=your_twilio_token
GMAIL_APP_PASSWORD=your_gmail_app_password

# Smart Home
HOMEASSISTANT_URL=http://homeassistant.local:8123
HOMEASSISTANT_TOKEN=your_ha_token

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/focus_helper
```

## 🚀 Getting Started

### Step 1: Set Up n8n
1. Install n8n on your server
2. Create a new workflow
3. Add a webhook trigger node
4. Configure the webhook URL

### Step 2: Configure Focus Helper
1. Update your profile with the n8n webhook URL
2. Enable webhook notifications
3. Test the connection

### Step 3: Test the Integration
1. Trigger a test alert in Focus Helper
2. Verify the webhook is received in n8n
3. Check that notifications are sent correctly

### Step 4: Customize for Your Needs
1. Adjust alert levels and timing
2. Add your own notification methods
3. Integrate with your existing systems

## 📞 Support

For help with n8n integration:
- Check the [n8n documentation](https://docs.n8n.io/)
- Join the Focus Helper community forum
- Contact support for custom integrations

This integration system provides a comprehensive safety net while respecting your autonomy and independence. It's designed to support you, not control you.
