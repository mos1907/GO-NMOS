-- Add bucket_id to flows table for Planner integration
-- This allows flows to be assigned to address buckets (drives/folders/views)

ALTER TABLE flows ADD COLUMN IF NOT EXISTS bucket_id BIGINT REFERENCES address_buckets(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_flows_bucket_id ON flows(bucket_id);
