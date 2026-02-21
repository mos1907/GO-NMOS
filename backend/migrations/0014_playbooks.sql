-- E.1: Operational playbooks - reusable workflows for TV campus operations
CREATE TABLE IF NOT EXISTS playbooks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,  -- Array of action steps
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Parameter definitions (name, type, description)
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_playbooks_enabled ON playbooks(enabled);

-- Playbook execution history
CREATE TABLE IF NOT EXISTS playbook_executions (
    id BIGSERIAL PRIMARY KEY,
    playbook_id TEXT NOT NULL REFERENCES playbooks(id) ON DELETE CASCADE,
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Actual parameter values used
    status TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'success', 'error')),
    result JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Execution result/error details
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_playbook_executions_playbook_id ON playbook_executions(playbook_id);
CREATE INDEX IF NOT EXISTS idx_playbook_executions_started_at ON playbook_executions(started_at DESC);

-- Example playbooks (can be customized)
INSERT INTO playbooks(id, name, description, steps, parameters)
VALUES
(
    'failover_backup_encoder',
    'Failover to Backup Encoder',
    'Disconnect current encoder receiver and connect backup encoder receiver for a channel',
    '[
        {
            "action": "disconnect_receiver",
            "receiver_id": "{{primary_receiver_id}}",
            "description": "Disconnect primary encoder receiver"
        },
        {
            "action": "connect_receiver",
            "receiver_id": "{{backup_receiver_id}}",
            "sender_id": "{{sender_id}}",
            "description": "Connect backup encoder receiver"
        }
    ]'::jsonb,
    '{
        "primary_receiver_id": {"type": "string", "description": "Primary receiver ID to disconnect", "required": true},
        "backup_receiver_id": {"type": "string", "description": "Backup receiver ID to connect", "required": true},
        "sender_id": {"type": "string", "description": "Sender ID to connect backup receiver to", "required": true}
    }'::jsonb
),
(
    'swap_studio_ab',
    'Swap Studio A/B Control',
    'Swap receivers between Studio A and Studio B',
    '[
        {
            "action": "disconnect_receiver",
            "receiver_id": "{{studio_a_receiver_id}}",
            "description": "Disconnect Studio A receiver"
        },
        {
            "action": "disconnect_receiver",
            "receiver_id": "{{studio_b_receiver_id}}",
            "description": "Disconnect Studio B receiver"
        },
        {
            "action": "connect_receiver",
            "receiver_id": "{{studio_a_receiver_id}}",
            "sender_id": "{{studio_b_sender_id}}",
            "description": "Connect Studio A receiver to Studio B sender"
        },
        {
            "action": "connect_receiver",
            "receiver_id": "{{studio_b_receiver_id}}",
            "sender_id": "{{studio_a_sender_id}}",
            "description": "Connect Studio B receiver to Studio A sender"
        }
    ]'::jsonb,
    '{
        "studio_a_receiver_id": {"type": "string", "description": "Studio A receiver ID", "required": true},
        "studio_b_receiver_id": {"type": "string", "description": "Studio B receiver ID", "required": true},
        "studio_a_sender_id": {"type": "string", "description": "Studio A sender ID", "required": true},
        "studio_b_sender_id": {"type": "string", "description": "Studio B sender ID", "required": true}
    }'::jsonb
),
(
    'toggle_test_chain',
    'Enable/Disable Test Chain',
    'Enable or disable a test chain on an output receiver',
    '[
        {
            "action": "{{action}}",
            "receiver_id": "{{receiver_id}}",
            "sender_id": "{{sender_id}}",
            "description": "{{action}} test chain on output {{receiver_id}}"
        }
    ]'::jsonb,
    '{
        "action": {"type": "string", "description": "Action: connect_receiver or disconnect_receiver", "required": true, "enum": ["connect_receiver", "disconnect_receiver"]},
        "receiver_id": {"type": "string", "description": "Test chain receiver ID", "required": true},
        "sender_id": {"type": "string", "description": "Sender ID (required if action is connect_receiver)", "required": false}
    }'::jsonb
)
ON CONFLICT (id) DO NOTHING;
