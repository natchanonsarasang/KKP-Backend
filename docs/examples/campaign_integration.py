# -*- coding: utf-8 -*-
"""
Callecto API - Outbound Campaign Integration Example (Python)

Requires the `requests` package:
    pip install requests
"""

import requests
from datetime import datetime, timezone, timedelta

BASE_URL = "http://localhost:8080"
JWT_TOKEN = "YOUR_JWT_TOKEN"

headers = {
    "Authorization": f"Bearer {JWT_TOKEN}",
    "Content-Type": "application/json"
}

def create_calling_session(workspace_id):
    """
    Creates a campaign calling session config setting concurrent call counts and delays.
    """
    url = f"{BASE_URL}/api/v1/call-sessions"
    
    payload = {
        "workspace_id": workspace_id,
        "status": "pending",
        "settings": {
            "maxRetries": 3,
            "delayBetweenCalls": 15,
            "concurrentCalls": 2,
            "businessHoursOnly": True,
            "businessHoursStart": "09:00",
            "businessHoursEnd": "17:00",
            "businessDays": [1, 2, 3, 4, 5],
            "testMode": False,
            "timezoneOffset": 420,  # +07:00 (ICT Bangkok)
            "interruptible": True
        }
    }
    
    response = requests.post(url, headers=headers, json=payload)
    if response.status_code != 200:
        raise Exception(f"Create Session Failed: {response.text}")
    
    print("Call Campaign Session Configured successfully.")
    return response.json()

def check_campaign_attempts(workspace_id):
    """
    Queries attempt logs within a workspace.
    """
    url = f"{BASE_URL}/api/v1/call-attempts/workspace/{workspace_id}"
    params = {
        "limit": 10,
        "status": "finished"
    }
    
    response = requests.get(url, headers=headers, params=params)
    if response.status_code != 200:
        raise Exception(f"Fetch Attempts Failed: {response.text}")
        
    attempts = response.json()
    print(f"Retrieved {len(attempts.get('data', []))} finished call attempts.")
    return attempts

if __name__ == "__main__":
    try:
        # Example variables
        demo_workspace_id = "wsp_92831"
        create_calling_session(demo_workspace_id)
        check_campaign_attempts(demo_workspace_id)
    except Exception as e:
        print(f"Error during execution: {e}")
