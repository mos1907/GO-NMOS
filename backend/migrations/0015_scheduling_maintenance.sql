-- E.2: Scheduling & maintenance windows - time-based playbook execution and maintenance periods

-- Scheduled playbook executions
CREATE TABLE IF NOT EXISTS scheduled_playbook_executions (
    id BIGSERIAL PRIMARY KEY,
    playbook_id TEXT NOT NULL REFERENCES playbooks(id) ON DELETE CASCADE,
    parameters JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Parameter values for this execution
    scheduled_at TIMESTAMPTZ NOT NULL,  -- When to execute
    executed_at TIMESTAMPTZ,  -- When it was executed (NULL = pending)
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'executed', 'failed', 'cancelled')),
    execution_id BIGINT,  -- Link to playbook_executions table if executed
    created_by TEXT,  -- Username
    result JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Execution result
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduled_playbook_executions_scheduled_at ON scheduled_playbook_executions(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_scheduled_playbook_executions_status ON scheduled_playbook_executions(status);
CREATE INDEX IF NOT EXISTS idx_scheduled_playbook_executions_playbook_id ON scheduled_playbook_executions(playbook_id);

-- Maintenance windows - periods when different routing policies apply
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    routing_policy_id BIGINT REFERENCES routing_policies(id) ON DELETE SET NULL,  -- Policy to apply during maintenance
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by TEXT,  -- Username
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_maintenance_windows_start_time ON maintenance_windows(start_time);
CREATE INDEX IF NOT EXISTS idx_maintenance_windows_end_time ON maintenance_windows(end_time);
CREATE INDEX IF NOT EXISTS idx_maintenance_windows_enabled ON maintenance_windows(enabled) WHERE enabled = TRUE;
