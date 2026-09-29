package configstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const SchemaVersion = 1

const schemaV1 = `
CREATE TABLE IF NOT EXISTS configstore_schema (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    version integer NOT NULL CHECK (version > 0)
);
CREATE TABLE IF NOT EXISTS config_revisions (
    revision bigint PRIMARY KEY CHECK (revision > 0),
    payload jsonb NOT NULL,
    payload_sha256 text NOT NULL CHECK (length(payload_sha256) = 64),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by text NOT NULL
);
CREATE TABLE IF NOT EXISTS node_status (
    node_id uuid PRIMARY KEY,
    node_role text NOT NULL CHECK (node_role IN ('control-plane', 'data-plane', 'certificate-controller')),
    config_revision bigint NOT NULL DEFAULT 0,
    applied_revision bigint NOT NULL DEFAULT 0,
    last_error text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (applied_revision <= config_revision)
);
CREATE TABLE IF NOT EXISTS certificate_revisions (
    certificate_id uuid NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0),
    metadata jsonb NOT NULL,
    encrypted_key bytea NOT NULL,
    certificate_pem bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (certificate_id, revision)
);
CREATE TABLE IF NOT EXISTS controller_leases (
    lease_name text PRIMARY KEY,
    holder_id uuid NOT NULL,
    valid_until timestamptz NOT NULL,
    CHECK (lease_name = 'certificate-controller')
);
INSERT INTO configstore_schema (singleton, version) VALUES (true, 1)
ON CONFLICT (singleton) DO NOTHING;
`

// ApplySchema applies idempotent control-plane migrations in one transaction.
// A newer schema is never downgraded implicitly.
func ApplySchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("postgres database is required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, schemaV1); err != nil {
		return fmt.Errorf("apply configstore schema version %d: %w", SchemaVersion, err)
	}
	if err := verifySchemaVersion(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func verifySchemaVersion(ctx context.Context, db queryRower) error {
	var version int
	if err := db.QueryRowContext(ctx, `SELECT version FROM configstore_schema WHERE singleton = true`).Scan(&version); err != nil {
		return fmt.Errorf("read configstore schema version: %w", err)
	}
	if version != SchemaVersion {
		return fmt.Errorf("unsupported configstore schema version %d (expected %d)", version, SchemaVersion)
	}
	return nil
}

// VerifySchema checks connectivity and the exact schema version without
// mutating the database.
func VerifySchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("postgres database is required")
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to configstore postgres: %w", err)
	}
	return verifySchemaVersion(ctx, db)
}
