-- NMOS registry schema (IS-04 oriented) designed to be extensible for IS-05/08/09/06

CREATE TABLE IF NOT EXISTS nmos_nodes (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    api_version TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '{}'::jsonb,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nmos_devices (
    id TEXT PRIMARY KEY,
    node_id TEXT NOT NULL REFERENCES nmos_nodes(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '{}'::jsonb,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nmos_flows (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT '',
    source_id TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '{}'::jsonb,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nmos_senders (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES nmos_devices(id) ON DELETE CASCADE,
    flow_id TEXT NOT NULL REFERENCES nmos_flows(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    transport TEXT NOT NULL DEFAULT '',
    manifest_href TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '{}'::jsonb,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nmos_receivers (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES nmos_devices(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT '',
    transport TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '{}'::jsonb,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nmos_devices_node_id ON nmos_devices(node_id);
CREATE INDEX IF NOT EXISTS idx_nmos_senders_device_id ON nmos_senders(device_id);
CREATE INDEX IF NOT EXISTS idx_nmos_senders_flow_id ON nmos_senders(flow_id);
CREATE INDEX IF NOT EXISTS idx_nmos_receivers_device_id ON nmos_receivers(device_id);

