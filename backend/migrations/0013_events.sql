-- C.3: IS-07 / Event & Tally: store events for routing (tally, GPI, alarms) and correlation with flows/senders/receivers

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    source_url TEXT NOT NULL DEFAULT '',   -- IS-07 device or source URL
    source_id TEXT NOT NULL DEFAULT '',    -- Event source ID (e.g. IS-07 source id)
    severity TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,  -- Full event payload (type, value, etc.)
    flow_id TEXT,    -- NMOS flow ID if correlated
    sender_id TEXT, -- NMOS sender ID if correlated
    receiver_id TEXT,-- NMOS receiver ID if correlated
    job_id TEXT,     -- Automation job ID if correlated
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_source ON events(source_url, source_id);
CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity);
CREATE INDEX IF NOT EXISTS idx_events_flow_id ON events(flow_id) WHERE flow_id IS NOT NULL AND flow_id != '';
CREATE INDEX IF NOT EXISTS idx_events_sender_id ON events(sender_id) WHERE sender_id IS NOT NULL AND sender_id != '';
CREATE INDEX IF NOT EXISTS idx_events_receiver_id ON events(receiver_id) WHERE receiver_id IS NOT NULL AND receiver_id != '';
