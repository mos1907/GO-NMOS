-- B.1 Connection state model: track IS-05 receiver connection state (staged/active, master/backup)

CREATE TABLE IF NOT EXISTS receiver_connections (
    id BIGSERIAL PRIMARY KEY,
    receiver_id TEXT NOT NULL, -- NMOS receiver ID (IS-04)
    state TEXT NOT NULL CHECK (state IN ('staged', 'active')), -- IS-05 connection state
    role TEXT NOT NULL DEFAULT 'master' CHECK (role IN ('master', 'backup')), -- master vs backup flow
    sender_id TEXT NOT NULL DEFAULT '', -- NMOS sender ID (IS-04)
    flow_id BIGINT REFERENCES flows(id) ON DELETE SET NULL, -- Optional link to internal flow record
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by TEXT, -- Username who made the change
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb, -- Extensible: transport_params, constraints, etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(receiver_id, state, role) -- One staged and one active per receiver, per role
);

CREATE INDEX IF NOT EXISTS idx_receiver_connections_receiver_id ON receiver_connections(receiver_id);
CREATE INDEX IF NOT EXISTS idx_receiver_connections_state ON receiver_connections(state);
CREATE INDEX IF NOT EXISTS idx_receiver_connections_flow_id ON receiver_connections(flow_id);
CREATE INDEX IF NOT EXISTS idx_receiver_connections_changed_at ON receiver_connections(changed_at DESC);

-- Connection history table for audit trail (keeps last N changes per receiver)
CREATE TABLE IF NOT EXISTS receiver_connection_history (
    id BIGSERIAL PRIMARY KEY,
    receiver_id TEXT NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('staged', 'active')),
    role TEXT NOT NULL DEFAULT 'master' CHECK (role IN ('master', 'backup')),
    sender_id TEXT NOT NULL DEFAULT '',
    flow_id BIGINT REFERENCES flows(id) ON DELETE SET NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by TEXT,
    action TEXT NOT NULL DEFAULT 'connect' CHECK (action IN ('connect', 'disconnect', 'update')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receiver_connection_history_receiver_id ON receiver_connection_history(receiver_id, changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_receiver_connection_history_changed_at ON receiver_connection_history(changed_at DESC);
