-- Create replication user for Postgres replica (ignore if exists)
DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'replicator') THEN
      CREATE USER replicator WITH REPLICATION LOGIN ENCRYPTED PASSWORD 'replicator_password';
   END IF;
END
$$;

-- Create replication slots (ignore if exists)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'replication_slot')
    THEN pg_create_physical_replication_slot('replication_slot')
    ELSE NULL
END;

SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'logical_slot')
    THEN (SELECT * FROM pg_create_logical_replication_slot('logical_slot', 'pgoutput'))
    ELSE NULL
END;

-- Create databases for CRM testing (conditional)
SELECT 'CREATE DATABASE crm_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'crm_test')\gexec

SELECT 'CREATE DATABASE transaction'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'transaction')\gexec

SELECT 'CREATE DATABASE onboarding'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'onboarding')\gexec

-- Connect to onboarding database and create tables
\c onboarding;

-- Migration 000000: Create organization table
CREATE TABLE IF NOT EXISTS organization
(
    id                                   UUID PRIMARY KEY NOT NULL,
    parent_organization_id               UUID,
    legal_name                           TEXT NOT NULL,
    doing_business_as                    TEXT,
    legal_document                       TEXT NOT NULL,
    address                              JSONB NOT NULL,
    status                               TEXT NOT NULL,
    status_description                   TEXT,
    created_at                           TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                           TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (parent_organization_id) REFERENCES organization (id)
);

CREATE INDEX idx_organization_created_at ON organization (created_at);

-- Migration 000001: Create ledger table
CREATE TABLE IF NOT EXISTS ledger
(
    id                            UUID PRIMARY KEY NOT NULL,
    name                          TEXT NOT NULL,
    organization_id               UUID NOT NULL,
    status                        TEXT NOT NULL,
    status_description            TEXT,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (organization_id) REFERENCES organization (id)
);

CREATE INDEX idx_ledger_created_at ON ledger (created_at);

-- Migration 000002: Create asset table
CREATE TABLE IF NOT EXISTS asset
(
    id                            UUID PRIMARY KEY NOT NULL,
    name                          TEXT,
    type                          TEXT NOT NULL,
    code                          TEXT NOT NULL,
    status                        TEXT NOT NULL,
    status_description            TEXT,
    ledger_id                     UUID NOT NULL,
    organization_id               UUID NOT NULL,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (ledger_id)       REFERENCES ledger (id),
    FOREIGN KEY (organization_id) REFERENCES organization (id)
);

CREATE INDEX idx_asset_search ON asset (organization_id, ledger_id, deleted_at, created_at DESC) INCLUDE (name, code);

-- Migration 000003: Create segment table
CREATE TABLE IF NOT EXISTS segment
(
    id                            UUID PRIMARY KEY NOT NULL,
    name                          TEXT,
    ledger_id                     UUID NOT NULL,
    organization_id               UUID NOT NULL,
    status                        TEXT NOT NULL,
    status_description            TEXT,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (ledger_id)       REFERENCES ledger (id),
    FOREIGN KEY (organization_id) REFERENCES organization (id)
);

CREATE INDEX idx_segment_created_at ON segment (created_at);

-- Migration 000004: Create portfolio table
CREATE TABLE IF NOT EXISTS portfolio
(
    id                            UUID PRIMARY KEY NOT NULL,
    name                          TEXT,
    entity_id                     TEXT NOT NULL,
    ledger_id                     UUID NOT NULL,
    organization_id               UUID NOT NULL,
    status                        TEXT NOT NULL,
    status_description            TEXT,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                    TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (ledger_id)       REFERENCES ledger (id),
    FOREIGN KEY (organization_id) REFERENCES organization (id)
);

CREATE INDEX idx_portfolio_id ON portfolio (organization_id, ledger_id, id, deleted_at, created_at);

-- Migration 000005: Create account table
CREATE TABLE IF NOT EXISTS account
(
    id                              UUID PRIMARY KEY NOT NULL,
    name                            TEXT,
    parent_account_id               UUID,
    entity_id                       TEXT,
    asset_code                      TEXT NOT NULL,
    organization_id                 UUID NOT NULL,
    ledger_id                       UUID NOT NULL,
    portfolio_id                    UUID,
    segment_id                      UUID,
    status                          TEXT NOT NULL,
    status_description              TEXT,
    alias                           TEXT NOT NULL,
    type                            TEXT NOT NULL,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (parent_account_id) REFERENCES account (id),
    FOREIGN KEY (organization_id)   REFERENCES organization (id),
    FOREIGN KEY (ledger_id)         REFERENCES ledger (id),
    FOREIGN KEY (portfolio_id)      REFERENCES portfolio (id),
    FOREIGN KEY (segment_id)        REFERENCES segment (id)
);

CREATE INDEX idx_account_alias ON account (organization_id, ledger_id, alias, deleted_at, created_at);

-- Migration 000006: Create account_type table
CREATE TABLE IF NOT EXISTS account_type (
    id                  UUID PRIMARY KEY NOT NULL,
    organization_id     UUID NOT NULL,
    ledger_id           UUID NOT NULL,
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    key_value           VARCHAR(50) NOT NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (organization_id) REFERENCES organization (id),
    FOREIGN KEY (ledger_id) REFERENCES ledger (id)
);

CREATE INDEX idx_account_type_organization_id_ledger_id ON account_type (organization_id, ledger_id);
CREATE INDEX idx_account_type_key_value ON account_type (organization_id, ledger_id, key_value) WHERE deleted_at IS NULL;
CREATE INDEX idx_account_type_deleted_at ON account_type (organization_id, ledger_id, deleted_at);
CREATE UNIQUE INDEX idx_account_type_unique_key_value ON account_type (organization_id, ledger_id, key_value) WHERE deleted_at IS NULL;

