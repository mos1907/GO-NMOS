-- Video/audio format summary for display (e.g. 1080i50, 1080p25, 720p50, L24/48k)
ALTER TABLE flows ADD COLUMN IF NOT EXISTS format_summary TEXT NOT NULL DEFAULT '';
