-- Migration 003: Service Accounts and API Keys
-- POC-004.0: Non-human identities

-- Service Accounts table
CREATE TABLE service_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    client_id VARCHAR(50) UNIQUE NOT NULL,
    client_secret_hash TEXT NOT NULL,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP
);

CREATE INDEX idx_service_accounts_client_id ON service_accounts(client_id);
CREATE INDEX idx_service_accounts_tenant_id ON service_accounts(tenant_id);

-- API Keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    key_prefix VARCHAR(20) NOT NULL,
    key_hash TEXT NOT NULL,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    expires_at TIMESTAMP,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP
);

CREATE UNIQUE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);

-- API Key Permissions junction table
CREATE TABLE api_key_permissions (
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (api_key_id, permission_id)
);

-- Service Account Roles junction table
CREATE TABLE service_account_roles (
    service_account_id UUID NOT NULL REFERENCES service_accounts(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (service_account_id, role_id)
);

-- New permissions for service accounts
INSERT INTO permissions (resource_type, action, description) VALUES
('bastion:service-account', 'create', 'Create service accounts'),
('bastion:service-account', 'read', 'View service account details'),
('bastion:service-account', 'update', 'Modify service accounts'),
('bastion:service-account', 'delete', 'Delete service accounts');

-- New permissions for API keys
INSERT INTO permissions (resource_type, action, description) VALUES
('bastion:api-key', 'create', 'Create API keys'),
('bastion:api-key', 'read', 'View API key details'),
('bastion:api-key', 'delete', 'Delete API keys');

-- Grant all service account and API key permissions to platform:superadmin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.role_type = 'platform' AND r.name = 'platform:superadmin'
  AND p.resource_type IN ('bastion:service-account', 'bastion:api-key');
