/**
 * Callecto API - Workspace & Debtor Lifecycle Integration Example (JavaScript/Node.js)
 * 
 * Shows how to:
 * 1. Create a Workspace
 * 2. Create a Debtor record within the Workspace
 * 3. Schedule a Campaign Queue (Call List Item)
 * 4. Trigger Campaign Execution
 */

const BASE_URL = 'http://localhost:8080';
const JWT_TOKEN = 'YOUR_JWT_TOKEN'; // Replace with a valid decoded token identifier

const headers = {
  'Authorization': `Bearer ${JWT_TOKEN}`,
  'Content-Type': 'application/json',
};

// 1. Create Workspace
async function createWorkspace(name) {
  const response = await fetch(`${BASE_URL}/api/v1/workspaces`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Create Workspace Failed: ${error.message}`);
  }

  const result = await response.json();
  console.log('Workspace Created Successfully:', result);
  return result;
}

// 2. Add Debtor
async function addDebtor(workspaceId, name, phoneNumber, debtAmount) {
  const debtorData = {
    name,
    phone_number: phoneNumber,
    total_debt: debtAmount,
    workspace_id: workspaceId,
    auto_call_enabled: true,
    status: 'active',
    variables: {
      due_amount: `${debtAmount} THB`,
      due_date: '2026-06-30',
    },
  };

  const response = await fetch(`${BASE_URL}/api/v1/debtors`, {
    method: 'POST',
    headers,
    body: JSON.stringify(debtorData),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Add Debtor Failed: ${error.message}`);
  }

  const result = await response.json();
  console.log('Debtor Added Successfully:', result);
  return result;
}

// 3. Queue Call List Item
async function queueCallListItem(workspaceId, debtorId, templateId) {
  const callListItem = {
    debtor_id: debtorId,
    workspace_id: workspaceId,
    template_id: templateId,
    status: 'pending',
    scheduled_at: new Date(Date.now() + 600000).toISOString(), // Schedule 10 mins from now
  };

  const response = await fetch(`${BASE_URL}/api/v1/call-list-items`, {
    method: 'POST',
    headers,
    body: JSON.stringify(callListItem),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Queue Call Failed: ${error.message}`);
  }

  const result = await response.json();
  console.log('Call Queue Item Enqueued Successfully:', result);
  return result;
}

// 4. Control Session Process Execution
async function controlSession(sessionId, action) {
  const response = await fetch(`${BASE_URL}/api/v1/call-process`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      session_id: sessionId,
      action: action, // 'start' | 'continue' | 'pause' | 'stop'
    }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Session Control Action Failed: ${error.message}`);
  }

  const result = await response.json();
  console.log(`Campaign Session Control '${action}' Triggered Successfully:`, result);
  return result;
}

// Main execution block orchestrating sequence
(async () => {
  try {
    console.log('Starting Callecto Integration Sequence...');
    const workspace = await createWorkspace('Automated Collection Q2');
    
    // Using placeholders for demonstrating flow logic
    const workspaceId = 'wsp_92831';
    const debtorId = 'deb_772819';
    const templateId = 'tpl_44928';
    const sessionId = '8fa19e34-bb3e-4361-bd80-d0f171050fb3';

    await addDebtor(workspaceId, 'Jane', '0812345678', 8400.0);
    await queueCallListItem(workspaceId, debtorId, templateId);
    await controlSession(sessionId, 'start');
  } catch (error) {
    console.error('Sequence Execution Error:', error.message);
  }
})();
