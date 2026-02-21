-- B.4: Optional SDN path association for flows (IS-06 / network controller)
ALTER TABLE flows ADD COLUMN IF NOT EXISTS sdn_path_id TEXT;
