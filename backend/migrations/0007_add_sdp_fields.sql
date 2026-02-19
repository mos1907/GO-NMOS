-- Add SDP (Session Description Protocol) fields to flows table
-- SDP is used in ST 2110 / NMOS for flow metadata (manifest_href from IS-04 senders)
-- sdp_url: URL to the SDP manifest (typically sender's manifest_href)
-- sdp_cache: Cached SDP text content for offline viewing and parsing

ALTER TABLE flows
ADD COLUMN IF NOT EXISTS sdp_url TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS sdp_cache TEXT NOT NULL DEFAULT '';
