-- B.3 Policy-based routing: rules for allowed/forbidden source-destination pairs, path requirements, constraints

CREATE TABLE IF NOT EXISTS routing_policies (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL, -- Human-friendly label, e.g. "No test feeds on TX outputs"
    policy_type TEXT NOT NULL CHECK (policy_type IN ('allowed_pair', 'forbidden_pair', 'path_requirement', 'constraint')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    -- Source/destination matching (can be wildcards or specific IDs)
    source_pattern TEXT, -- e.g. "sender:*", "flow:test-*", "device:TX-*"
    destination_pattern TEXT, -- e.g. "receiver:*", "device:TX-*"
    -- Path requirements (for path_requirement type)
    require_path_a BOOLEAN DEFAULT false,
    require_path_b BOOLEAN DEFAULT false,
    -- Constraint details (for constraint type)
    constraint_field TEXT, -- e.g. "format", "site", "room"
    constraint_value TEXT, -- e.g. "test", "TX", "Studio1"
    constraint_operator TEXT DEFAULT 'equals' CHECK (constraint_operator IN ('equals', 'contains', 'starts_with', 'ends_with')),
    -- Metadata
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 100, -- Lower = higher priority (checked first)
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routing_policies_enabled ON routing_policies(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_routing_policies_type ON routing_policies(policy_type);
CREATE INDEX IF NOT EXISTS idx_routing_policies_priority ON routing_policies(priority ASC);

-- Policy audit log: tracks policy violations and overrides
CREATE TABLE IF NOT EXISTS routing_policy_audit (
    id BIGSERIAL PRIMARY KEY,
    policy_id BIGINT REFERENCES routing_policies(id) ON DELETE SET NULL,
    action TEXT NOT NULL CHECK (action IN ('check', 'violation', 'override', 'allowed')),
    source_id TEXT, -- NMOS sender/flow ID
    destination_id TEXT, -- NMOS receiver ID
    flow_id BIGINT REFERENCES flows(id) ON DELETE SET NULL,
    violation_reason TEXT, -- Why policy was violated
    overridden_by TEXT, -- Username who overrode
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb, -- Additional context
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routing_policy_audit_policy_id ON routing_policy_audit(policy_id);
CREATE INDEX IF NOT EXISTS idx_routing_policy_audit_created_at ON routing_policy_audit(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_routing_policy_audit_action ON routing_policy_audit(action);
