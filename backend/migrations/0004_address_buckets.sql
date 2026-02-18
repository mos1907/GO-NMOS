CREATE TABLE IF NOT EXISTS address_buckets (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES address_buckets(id) ON DELETE CASCADE,
    bucket_type TEXT NOT NULL CHECK (bucket_type IN ('drive', 'parent', 'child')),
    name TEXT NOT NULL,
    cidr TEXT NOT NULL DEFAULT '',
    start_ip TEXT NOT NULL DEFAULT '',
    end_ip TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_address_buckets_parent_id ON address_buckets(parent_id);
CREATE INDEX IF NOT EXISTS idx_address_buckets_bucket_type ON address_buckets(bucket_type);

INSERT INTO address_buckets(parent_id, bucket_type, name, cidr, description)
VALUES
(NULL, 'drive', '239.0.0.0/8', '239.0.0.0/8', 'Default multicast drive')
ON CONFLICT DO NOTHING;
