-- POC-003.0: Basic RBAC Schema Migration
-- Creates tenants, roles, permissions, and junction tables

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    role_type VARCHAR(50) NOT NULL CHECK (role_type IN ('platform', 'application')),
    application_name VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_type ON roles(role_type);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    UNIQUE(resource_type, action)
);

CREATE INDEX idx_permissions_resource ON permissions(resource_type);

-- Role-Permission junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role assignments (within tenants)
-- tenant_id is nullable for platform roles
-- Note: PostgreSQL allows NULL in composite primary keys but treats each NULL as distinct
-- We use a unique constraint instead to prevent duplicate assignments
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_roles_unique ON user_roles(user_id, role_id, COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid));

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_tenant ON user_roles(tenant_id);

-- Add tenant_id to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);

-- Seed platform roles
INSERT INTO roles (name, description, role_type) VALUES
('platform:superadmin', 'Full platform access', 'platform'),
('platform:admin', 'Platform administration', 'platform'),
('platform:auditor', 'Read-only audit access', 'platform')
ON CONFLICT (name) DO NOTHING;

-- Seed application roles for Bastion
INSERT INTO roles (name, description, role_type, application_name) VALUES
('bastion:tenant-admin', 'Manage tenant users and roles', 'application', 'bastion'),
('bastion:user-admin', 'Manage users within tenant', 'application', 'bastion'),
('bastion:viewer', 'Read-only access', 'application', 'bastion')
ON CONFLICT (name) DO NOTHING;

-- Seed permissions
INSERT INTO permissions (resource_type, action, description) VALUES
-- Tenant permissions
('bastion:tenant', 'create', 'Create new tenants'),
('bastion:tenant', 'read', 'View tenant details'),
('bastion:tenant', 'update', 'Modify tenant settings'),
('bastion:tenant', 'delete', 'Delete tenants'),
-- User permissions
('bastion:user', 'create', 'Create users'),
('bastion:user', 'read', 'View user details'),
('bastion:user', 'update', 'Modify users'),
('bastion:user', 'delete', 'Delete users'),
-- Role permissions
('bastion:role', 'assign', 'Assign roles to users'),
('bastion:role', 'revoke', 'Revoke roles from users'),
('bastion:role', 'read', 'View role definitions'),
-- Audit permissions
('bastion:audit', 'read', 'View audit logs')
ON CONFLICT (resource_type, action) DO NOTHING;

-- Assign permissions to platform:superadmin (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'platform:superadmin'
ON CONFLICT DO NOTHING;

-- Assign permissions to platform:admin (tenant, user, role management)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'platform:admin'
AND p.resource_type IN ('bastion:tenant', 'bastion:user', 'bastion:role')
ON CONFLICT DO NOTHING;

-- Assign permissions to platform:auditor (read-only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'platform:auditor'
AND p.action IN ('read')
ON CONFLICT DO NOTHING;

-- Assign permissions to bastion:tenant-admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'bastion:tenant-admin'
AND (
    (p.resource_type = 'bastion:user') OR
    (p.resource_type = 'bastion:role')
)
ON CONFLICT DO NOTHING;

-- Assign permissions to bastion:user-admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'bastion:user-admin'
AND p.resource_type = 'bastion:user'
ON CONFLICT DO NOTHING;

-- Assign permissions to bastion:viewer (read-only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'bastion:viewer'
AND p.action = 'read'
AND p.resource_type IN ('bastion:user', 'bastion:role')
ON CONFLICT DO NOTHING;

-- Bootstrap: Assign platform:superadmin to admin@bastion.local
-- This is idempotent - will only assign if user exists and doesn't have the role
INSERT INTO user_roles (user_id, role_id, tenant_id)
SELECT u.id, r.id, NULL
FROM users u
CROSS JOIN roles r
WHERE u.email = 'admin@bastion.local'
AND r.name = 'platform:superadmin'
AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.user_id = u.id AND ur.role_id = r.id
)
ON CONFLICT DO NOTHING;
