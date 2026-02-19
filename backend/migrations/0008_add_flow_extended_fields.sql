-- Extend flows table to support ST2022-7 A/B paths and richer NMOS metadata (mmam-docker parity, backward compatible)
-- Keep existing single-path fields (multicast_ip/source_ip/port) for compatibility.

ALTER TABLE flows
-- ST2022-7 A/B paths
ADD COLUMN IF NOT EXISTS source_addr_a TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS source_port_a INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS multicast_addr_a TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS group_port_a INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS source_addr_b TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS source_port_b INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS multicast_addr_b TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS group_port_b INTEGER NOT NULL DEFAULT 0,
-- NMOS metadata identifiers
ADD COLUMN IF NOT EXISTS nmos_node_id UUID,
ADD COLUMN IF NOT EXISTS nmos_flow_id UUID,
ADD COLUMN IF NOT EXISTS nmos_sender_id UUID,
ADD COLUMN IF NOT EXISTS nmos_device_id UUID,
-- NMOS node text metadata
ADD COLUMN IF NOT EXISTS nmos_node_label TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_node_description TEXT NOT NULL DEFAULT '',
-- NMOS endpoint information
ADD COLUMN IF NOT EXISTS nmos_is04_host TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_is04_port INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS nmos_is05_host TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_is05_port INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS nmos_is04_base_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_is05_base_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_is04_version TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_is05_version TEXT NOT NULL DEFAULT '',
-- Label/description/management
ADD COLUMN IF NOT EXISTS nmos_label TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS nmos_description TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS management_url TEXT NOT NULL DEFAULT '',
-- Media info (often derived from SDP/IS-05)
ADD COLUMN IF NOT EXISTS media_type TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS st2110_format TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS redundancy_group TEXT NOT NULL DEFAULT '',
-- Data source tracking
ADD COLUMN IF NOT EXISTS data_source TEXT NOT NULL DEFAULT 'manual',
ADD COLUMN IF NOT EXISTS rds_address TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS rds_api_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS rds_version TEXT NOT NULL DEFAULT '';

-- Helpful indexes (kept minimal to avoid migration surprises)
CREATE INDEX IF NOT EXISTS idx_flows_multicast_addr_a ON flows(multicast_addr_a) WHERE multicast_addr_a <> '';
CREATE INDEX IF NOT EXISTS idx_flows_group_port_a ON flows(group_port_a) WHERE group_port_a > 0;
CREATE INDEX IF NOT EXISTS idx_flows_nmos_flow_id ON flows(nmos_flow_id) WHERE nmos_flow_id IS NOT NULL;
