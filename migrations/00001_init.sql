-- +goose Up
create table tenants (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    slug text not null unique check (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    created_at timestamptz not null default now()
);

create table users (
    id uuid primary key default gen_random_uuid(),
    email text not null,
    name text,
    created_at timestamptz not null default now()
);

create unique index user_email_key on users (lower(email));

create table memberships (
    tenant_id uuid not null references tenants (id),
    user_id uuid not null references users (id),
    role text not null check (role in ('owner', 'admin', 'staff')),
    created_at timestamptz not null default now(),
    primary key (tenant_id, user_id)
);

create index usr_membership_idx on memberships (user_id);

-- +goose Down
drop index usr_membership_idx;
drop table memberships;
drop index user_email_key;
drop table users;
drop table tenants;
