#!/usr/bin/env python3
"""
MCP Client for Focus Helper
A Python client to interact with the Focus Helper MCP server
"""

import requests
import json
import time
from datetime import datetime, timedelta

class MCPClient:
    def __init__(self, base_url="http://localhost:8089"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})
    
    def _send_request(self, method, params=None):
        """Send a JSON-RPC request to the MCP server"""
        payload = {
            "jsonrpc": "2.0",
            "id": int(time.time() * 1000),
            "method": method,
            "params": params or {}
        }
        
        try:
            response = self.session.post(f"{self.base_url}/mcp", json=payload)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"❌ Request failed: {e}")
            return None
    
    def ping(self):
        """Test connection to the MCP server"""
        result = self._send_request("ping")
        if result and "result" in result:
            print(f"✅ Connected to MCP server: {result['result'].get('pong', 'unknown')}")
            return True
        else:
            print("❌ Failed to connect to MCP server")
            return False
    
    def get_session_info(self):
        """Get current session information"""
        result = self._send_request("get_session_info")
        if result and "result" in result:
            return result["result"]
        return None
    
    def get_alert_levels(self):
        """Get available alert levels"""
        result = self._send_request("get_alert_levels")
        if result and "result" in result:
            return result["result"]
        return None
    
    def trigger_alert(self, alert_index):
        """Trigger a specific alert level"""
        params = {"alert_index": alert_index}
        result = self._send_request("trigger_alert", params)
        if result and "result" in result:
            return result["result"]
        return None
    
    def get_hyperfocus_status(self):
        """Get current hyperfocus status"""
        result = self._send_request("get_hyperfocus_status")
        if result and "result" in result:
            return result["result"]
        return None

def format_duration(nanoseconds):
    """Convert nanoseconds to human-readable duration"""
    if nanoseconds == 0:
        return "0s"
    
    seconds = nanoseconds / 1_000_000_000
    if seconds < 60:
        return f"{seconds:.0f}s"
    elif seconds < 3600:
        return f"{seconds/60:.0f}m"
    else:
        return f"{seconds/3600:.1f}h"

def main():
    print("🤖 Focus Helper MCP Client (Python)")
    print("=" * 50)
    
    # Create MCP client
    client = MCPClient()
    
    # Test connection
    if not client.ping():
        return
    
    print("\n📊 Current Session Information:")
    session_info = client.get_session_info()
    if session_info:
        print(f"Session ID: {session_info.get('session_id', 'N/A')}")
        print(f"Subject: {session_info.get('subject', 'Unknown')}")
        print(f"Start Time: {session_info.get('start_time', 'N/A')}")
        print(f"Current Time: {session_info.get('current_time', 'N/A')}")
        print(f"Duration: {format_duration(session_info.get('duration', 0))}")
        print(f"Is Active: {session_info.get('is_active', False)}")
        print(f"Last Activity: {session_info.get('last_activity_time', 'N/A')}")
        print(f"Idle Duration: {format_duration(session_info.get('idle_duration', 0))}")
        
        if session_info.get('hyperfocus_level'):
            print(f"Hyperfocus Level: {session_info.get('hyperfocus_level')}")
            print(f"Hyperfocus Start: {session_info.get('hyperfocus_start_time', 'N/A')}")
            print(f"Hyperfocus Duration: {format_duration(session_info.get('hyperfocus_duration', 0))}")
    else:
        print("❌ Failed to get session information")
    
    print("\n🚨 Available Alert Levels:")
    alert_levels = client.get_alert_levels()
    if alert_levels:
        for level in alert_levels:
            threshold = format_duration(level.get('threshold', 0))
            print(f"Index {level.get('index', 0)}: {level.get('level', 'unknown')} "
                  f"(Enabled: {level.get('enabled', False)}, Threshold: {threshold})")
    else:
        print("❌ Failed to get alert levels")
    
    print("\n🎯 Hyperfocus Status:")
    hyperfocus_status = client.get_hyperfocus_status()
    if hyperfocus_status:
        for key, value in hyperfocus_status.items():
            if key == "duration" and isinstance(value, (int, float)):
                print(f"{key}: {format_duration(value)}")
            else:
                print(f"{key}: {value}")
    else:
        print("❌ Failed to get hyperfocus status")
    
    # Example: Trigger first available alert level
    if alert_levels and len(alert_levels) > 0:
        first_alert = alert_levels[0]
        print(f"\n🔔 Triggering alert level {first_alert.get('index', 0)} ({first_alert.get('level', 'unknown')})...")
        
        result = client.trigger_alert(first_alert.get('index', 0))
        if result:
            print(f"✅ {result.get('message', 'Alert triggered')}")
        else:
            print("❌ Failed to trigger alert")
    
    print("\n✅ MCP client demo completed")

if __name__ == "__main__":
    main()
