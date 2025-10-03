CREATE TABLE IF NOT EXISTS "transaction" (
    id                                  UUID PRIMARY KEY NOT NULL,
    parent_transaction_id               UUID,
    description                         TEXT NOT NULL,
    template                            TEXT NOT NULL,
    status                              TEXT NOT NULL,
    status_description                  TEXT,
    amount                              BIGINT NOT NULL,
    amount_scale                        BIGINT NOT NULL,
    asset_code                          TEXT NOT NULL,
    chart_of_accounts_group_name        TEXT NOT NULL,
    ledger_id                           UUID NOT NULL,
    organization_id                     UUID NOT NULL,
    body                                JSONB NOT NULL,
    created_at                          TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                          TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (parent_transaction_id) REFERENCES "transaction" (id)
);

CREATE INDEX idx_transaction_parent_transaction_id ON "transaction" (parent_transaction_id);
REINDEX INDEX idx_transaction_parent_transaction_id;

CREATE INDEX idx_transaction_organization_ledger_id ON "transaction" (organization_id, ledger_id);
REINDEX INDEX idx_transaction_organization_ledger_id;

CREATE INDEX idx_transaction_created_at ON "transaction" (created_at);
REINDEX INDEX idx_transaction_created_at;
CREATE TABLE IF NOT EXISTS operation (
    id                                 UUID PRIMARY KEY NOT NULL,
    transaction_id                     UUID NOT NULL,
    description                        TEXT NOT NULL,
    type                               TEXT NOT NULL,
    asset_code                         TEXT NOT NULL,
    amount                             BIGINT NOT NULL DEFAULT 0,
    amount_scale                       BIGINT NOT NULL DEFAULT 0,
    available_balance                  BIGINT NOT NULL DEFAULT 0,
    on_hold_balance                    BIGINT NOT NULL DEFAULT 0,
    balance_scale                      BIGINT NOT NULL DEFAULT 0,
    available_balance_after            BIGINT NOT NULL DEFAULT 0,
    on_hold_balance_after              BIGINT NOT NULL DEFAULT 0,
    balance_scale_after                BIGINT NOT NULL DEFAULT 0,
    status                             TEXT NOT NULL,
    status_description                 TEXT NULL,
    account_id                         UUID NOT NULL,
    account_alias                      TEXT NOT NULL,
    balance_id                         UUID NOT NULL,
    chart_of_accounts                  TEXT NOT NULL,
    organization_id                    UUID NOT NULL,
    ledger_id                          UUID NOT NULL,
    created_at                         TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                         TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (transaction_id) REFERENCES "transaction" (id)
);

CREATE INDEX idx_operation_organization_transaction_id ON operation (transaction_id);
REINDEX INDEX idx_operation_organization_transaction_id;

CREATE INDEX idx_operation_organization_ledger_id ON operation (organization_id, ledger_id);
REINDEX INDEX idx_operation_organization_ledger_id;

CREATE INDEX idx_operation_created_at ON operation (created_at);
REINDEX INDEX idx_operation_created_at;
CREATE TABLE IF NOT EXISTS asset_rate (
    id                                  UUID PRIMARY KEY NOT NULL,
    organization_id                     UUID NOT NULL,
    ledger_id                           UUID NOT NULL,
    external_id                         UUID NOT NULL,
    "from"                              TEXT NOT NULL,
    "to"                                TEXT NOT NULL,
    rate                                BIGINT NOT NULL,
    rate_scale                          NUMERIC NOT NULL,
    source                              TEXT,
    ttl                                 BIGINT NOT NULL,
    created_at                          TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_asset_rate_organization_ledger_id ON asset_rate (organization_id, ledger_id);
REINDEX INDEX idx_asset_rate_organization_ledger_id;

CREATE INDEX idx_asset_rate_created_at ON asset_rate (created_at);
REINDEX INDEX idx_asset_rate_created_at;
CREATE TABLE IF NOT EXISTS balance (
  id                                  UUID PRIMARY KEY NOT NULL,
  organization_id                     UUID NOT NULL,
  ledger_id                           UUID NOT NULL,
  account_id                          UUID NOT NULL,
  alias                               TEXT NOT NULL,
  asset_code                          TEXT NOT NULL,
  available                           BIGINT NOT NULL DEFAULT 0,
  on_hold                             BIGINT NOT NULL DEFAULT 0,
  scale                               BIGINT NOT NULL DEFAULT 0,
  version                             BIGINT DEFAULT 0,
  account_type                        TEXT NOT NULL,
  allow_sending                       BOOLEAN NOT NULL,
  allow_receiving                     BOOLEAN NOT NULL,
  created_at                          TIMESTAMP WITH TIME ZONE NOT NULL,
  updated_at                          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  deleted_at                          TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_balance_account_id ON balance (organization_id, ledger_id, account_id, deleted_at, created_at);
REINDEX INDEX idx_balance_account_id;

CREATE INDEX idx_balance_alias ON balance (organization_id, ledger_id, alias, deleted_at, created_at);
REINDEX INDEX idx_balance_alias;
BEGIN;

ALTER TABLE transaction
    ALTER COLUMN amount TYPE DECIMAL USING (amount / POWER(10, amount_scale::INTEGER))::DECIMAL;

COMMIT;

ALTER TABLE transaction
    DROP COLUMN IF EXISTS amount_scale;

ALTER TABLE transaction
    DROP COLUMN IF EXISTS template;

ALTER TABLE transaction
    ADD COLUMN IF NOT EXISTS route TEXT NULL;

ALTER TABLE transaction 
    ALTER COLUMN body DROP NOT NULL;

UPDATE transaction
    SET body = NULL;

VACUUM FULL transaction;
BEGIN;

ALTER TABLE balance
  ALTER COLUMN available TYPE DECIMAL USING (available / POWER(10, scale::INTEGER))::DECIMAL,
  ALTER COLUMN on_hold TYPE DECIMAL USING on_hold::DECIMAL;

COMMIT;

ALTER TABLE balance
    DROP COLUMN IF EXISTS scale;
BEGIN;

ALTER TABLE operation
  ALTER COLUMN amount TYPE DECIMAL USING (amount / POWER(10, amount_scale::INTEGER))::DECIMAL,
  ALTER COLUMN available_balance TYPE DECIMAL USING (available_balance / POWER(10, balance_scale::INTEGER))::DECIMAL,
  ALTER COLUMN on_hold_balance TYPE DECIMAL USING on_hold_balance::DECIMAL,
  ALTER COLUMN available_balance_after TYPE DECIMAL USING (available_balance_after / POWER(10, balance_scale_after::INTEGER))::DECIMAL,
  ALTER COLUMN on_hold_balance_after TYPE DECIMAL USING on_hold_balance_after::DECIMAL;

COMMIT;

ALTER TABLE operation
    DROP COLUMN IF EXISTS amount_scale,
    DROP COLUMN IF EXISTS balance_scale,
    DROP COLUMN IF EXISTS balance_scale_after;

ALTER TABLE operation
    ADD COLUMN IF NOT EXISTS route TEXT NULL;
CREATE TABLE IF NOT EXISTS operation_route (
    id                              UUID PRIMARY KEY NOT NULL,
    organization_id                 UUID NOT NULL,
    ledger_id                       UUID NOT NULL,
    title                           VARCHAR(255) NOT NULL,
    description                     VARCHAR(250),
    operation_type                  VARCHAR(20) NOT NULL CHECK (LOWER(operation_type) IN ('source', 'destination')),
    account_rule_type               VARCHAR(20),
    account_rule_valid_if           TEXT,
    created_at                      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_operation_route_organization_id_ledger_id ON operation_route (organization_id, ledger_id);

CREATE INDEX idx_operation_route_type ON operation_route (organization_id, ledger_id, operation_type) WHERE deleted_at IS NULL;

CREATE INDEX idx_operation_route_deleted_at ON operation_route (organization_id, ledger_id, deleted_at);
CREATE TABLE IF NOT EXISTS transaction_route (
    id                              UUID PRIMARY KEY NOT NULL,
    organization_id                 UUID NOT NULL,
    ledger_id                       UUID NOT NULL,
    title                           VARCHAR(255) NOT NULL,
    description                     VARCHAR(250),
    created_at                      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_transaction_route_organization_id_ledger_id ON transaction_route (organization_id, ledger_id);

CREATE TABLE IF NOT EXISTS operation_transaction_route (
    id UUID PRIMARY KEY NOT NULL,
    operation_route_id UUID NOT NULL REFERENCES operation_route(id),
    transaction_route_id UUID NOT NULL REFERENCES transaction_route(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_operation_transaction_route_unique 
ON operation_transaction_route (operation_route_id, transaction_route_id) 
WHERE deleted_at IS NULL;

CREATE INDEX idx_operation_transaction_route_operation_route_id
ON operation_transaction_route (operation_route_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_operation_transaction_route_transaction_route_id
ON operation_transaction_route (transaction_route_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_operation_transaction_route_deleted_at
ON operation_transaction_route (deleted_at);
ALTER TABLE operation
    ADD COLUMN IF NOT EXISTS balance_affected BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE balance ADD COLUMN key TEXT NOT NULL DEFAULT 'default';

CREATE INDEX idx_unique_balance_alias_key ON balance (organization_id, ledger_id, alias, key) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_balance_alias;
ALTER TABLE operation ADD COLUMN balance_key TEXT NOT NULL DEFAULT 'default';
ALTER TABLE operation_route ADD COLUMN code TEXT;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_operation_account 
ON operation (organization_id, ledger_id, account_id, deleted_at, created_at);
