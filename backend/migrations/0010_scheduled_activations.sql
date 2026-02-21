-- B.2 Scheduled activations: time-based IS-05 patch/take operations

CREATE TABLE IF NOT EXISTS scheduled_activations (
    id BIGSERIAL PRIMARY KEY,
    flow_id BIGINT NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
    receiver_ids TEXT[] NOT NULL, -- Array of NMOS receiver IDs
    is05_base_url TEXT NOT NULL,
    sender_id TEXT, -- Optional NMOS sender ID
    scheduled_at TIMESTAMPTZ NOT NULL, -- When to execute
    executed_at TIMESTAMPTZ, -- When it was executed (NULL = pending)
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'executed', 'failed', 'cancelled')),
    mode TEXT NOT NULL DEFAULT 'immediate' CHECK (mode IN ('immediate', 'safe_switch')),
    created_by TEXT, -- Username
    result JSONB NOT NULL DEFAULT '{}'::jsonb, -- Bulk patch result: {success, failed, results[]}
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduled_activations_scheduled_at ON scheduled_activations(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_scheduled_activations_status ON scheduled_activations(status);
CREATE INDEX IF NOT EXISTS idx_scheduled_activations_flow_id ON scheduled_activations(flow_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_activations_executed_at ON scheduled_activations(executed_at DESC);
