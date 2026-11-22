-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9-]+$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    is_system  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE roles
    ADD CONSTRAINT roles_tenant_name_uq
    UNIQUE NULLS NOT DISTINCT (tenant_id, name);

CREATE TABLE role_permissions (
    role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission TEXT NOT NULL, 
    PRIMARY KEY (role_id, permission)
);

CREATE TABLE memberships (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id)   ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, tenant_id)
);

CREATE INDEX memberships_tenant_idx ON memberships (tenant_id);
CREATE INDEX memberships_role_idx   ON memberships (role_id);

-- seed system roles
INSERT INTO roles (name, is_system) VALUES ('owner', true), ('member', true);

-- +goose Down
DROP TABLE memberships;
DROP TABLE role_permissions;
DROP TABLE roles;
DROP TABLE tenants;
DROP TABLE users;
