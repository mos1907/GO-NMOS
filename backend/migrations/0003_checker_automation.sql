CREATE TABLE IF NOT EXISTS checker_results (
    id BIGSERIAL PRIMARY KEY,
    kind TEXT NOT NULL,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_checker_results_kind_created_at
    ON checker_results(kind, created_at DESC);

CREATE TABLE IF NOT EXISTS scheduled_jobs (
    job_id TEXT PRIMARY KEY,
    job_type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    schedule_type TEXT NOT NULL DEFAULT 'interval',
    schedule_value TEXT NOT NULL DEFAULT '3600',
    last_run_at TIMESTAMPTZ,
    last_run_status TEXT,
    last_run_result JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO scheduled_jobs(job_id, job_type, enabled, schedule_type, schedule_value, last_run_status, last_run_result)
VALUES
('collision_check', 'collision', false, 'interval', '1800', 'idle', '{}'::jsonb),
('nmos_check', 'nmos', false, 'interval', '1800', 'idle', '{}'::jsonb)
ON CONFLICT (job_id) DO NOTHING;
