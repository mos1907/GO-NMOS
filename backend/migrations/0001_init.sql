CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flows (
    id BIGSERIAL PRIMARY KEY,
    flow_id UUID NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    multicast_ip TEXT NOT NULL DEFAULT '',
    source_ip TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    flow_status TEXT NOT NULL DEFAULT 'active',
    availability TEXT NOT NULL DEFAULT 'available',
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    note TEXT NOT NULL DEFAULT '',
    transport_protocol TEXT NOT NULL DEFAULT 'RTP/UDP',
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flows_updated_at ON flows(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_flows_multicast_ip ON flows(multicast_ip);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO settings(key, value) VALUES
('api_base_url', ''),
('anonymous_access', 'false'),
('flow_lock_role', 'editor'),
('hard_delete_enabled', 'false')
ON CONFLICT (key) DO NOTHING;
