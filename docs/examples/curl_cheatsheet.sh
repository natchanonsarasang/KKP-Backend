#!/usr/bin/env bash

# Callecto API cURL Cheatsheet
# Helper reference detailing routes and formats.

BASE_URL="http://localhost:8080"
JWT_TOKEN="YOUR_JWT_TOKEN"

echo "=== Get Workspaces ==="
curl -s -X GET "${BASE_URL}/api/v1/workspaces" \
  -H "Authorization: Bearer ${JWT_TOKEN}"

echo "=== Create Workspace ==="
curl -s -X POST "${BASE_URL}/api/v1/workspaces" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name": "Legal Department recoveries"}'

echo "=== Get Workspace Debtors ==="
curl -s -X GET "${BASE_URL}/api/v1/debtors/workspace/wsp_92831" \
  -H "Authorization: Bearer ${JWT_TOKEN}"

echo "=== Get Call list items inside Workspace ==="
curl -s -X GET "${BASE_URL}/api/v1/call-list-items/workspace/wsp_92831?statuses_in=pending,calling" \
  -H "Authorization: Bearer ${JWT_TOKEN}"

echo "=== Trigger Manual Voicebot Call ==="
curl -s -X POST "${BASE_URL}/api/v1/voicebot/make-call" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "0812345678",
    "variables": {"name": "John Doe", "amount": "12,000 THB"},
    "interruptible": true,
    "next_intent": "confirm_identity"
  }'

echo "=== Pause Calling Session ==="
curl -s -X POST "${BASE_URL}/api/v1/call-process" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "8fa19e34-bb3e-4361-bd80-d0f171050fb3",
    "action": "pause"
  }'
