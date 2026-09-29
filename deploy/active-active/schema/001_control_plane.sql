BEGIN;

CREATE TABLE IF NOT EXISTS config_revisions (
    revision       bigint PRIMARY KEY CHECK (revision > 0),
    payload        jsonb NOT NULL,
    payload_sha256 text NOT NULL CHECK (length(payload_sha256) = 64),
    created_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by     text NOT NULL
);

CREATE TABLE IF NOT EXISTS node_status (
    node_id          uuid PRIMARY KEY,
    node_role        text NOT NULL CHECK (node_role IN ('control-plane', 'data-plane', 'certificate-controller')),
    config_revision  bigint NOT NULL DEFAULT 0,
    applied_revision bigint NOT NULL DEFAULT 0,
    last_error       text NOT NULL DEFAULT '',
    updated_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (applied_revision <= config_revision)
);

CREATE TABLE IF NOT EXISTS certificate_revisions (
    certificate_id uuid NOT NULL,
    revision       bigint NOT NULL CHECK (revision > 0),
    metadata       jsonb NOT NULL,
    encrypted_key  bytea NOT NULL,
    certificate_pem bytea NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (certificate_id, revision)
);

CREATE TABLE IF NOT EXISTS controller_leases (
    lease_name  text PRIMARY KEY,
    holder_id   uuid NOT NULL,
    valid_until timestamptz NOT NULL,
    CHECK (lease_name = 'certificate-controller')
);

COMMIT;
