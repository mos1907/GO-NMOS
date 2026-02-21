-- E.3: Role-aware operations - strengthen role model and add playbook permissions

-- Update users table to support new roles: viewer, operator, engineer, admin, automation
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('viewer', 'operator', 'engineer', 'admin', 'automation'));

-- Add allowed_roles to playbooks table (which roles can execute this playbook)
ALTER TABLE playbooks ADD COLUMN IF NOT EXISTS allowed_roles TEXT[] NOT NULL DEFAULT ARRAY['engineer', 'admin']::TEXT[];

-- Update existing playbooks to allow engineer and admin by default
UPDATE playbooks SET allowed_roles = ARRAY['engineer', 'admin']::TEXT[] WHERE allowed_roles IS NULL OR array_length(allowed_roles, 1) IS NULL;

-- Create index for role-based queries
CREATE INDEX IF NOT EXISTS idx_playbooks_allowed_roles ON playbooks USING GIN(allowed_roles);
