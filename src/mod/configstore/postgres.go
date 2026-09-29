package configstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const revisionLockID int64 = 0x5a4f52415859 // "ZORAXY"

type PostgresStore struct {
	db           *sql.DB
	pollInterval time.Duration
}

func NewPostgresStore(db *sql.DB, pollInterval time.Duration) *PostgresStore {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &PostgresStore{db: db, pollInterval: pollInterval}
}

func scanRevision(row interface{ Scan(...any) error }) (Revision, error) {
	var revision Revision
	var payload []byte
	err := row.Scan(&revision.ID, &payload, &revision.SHA256, &revision.CreatedAt, &revision.CreatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return Revision{}, ErrNoRevision
	}
	if err != nil {
		return Revision{}, err
	}
	revision.Payload = json.RawMessage(payload)
	return revision, nil
}

func (s *PostgresStore) Current(ctx context.Context) (Revision, error) {
	return scanRevision(s.db.QueryRowContext(ctx, `
		SELECT revision, payload, payload_sha256, created_at, created_by
		FROM config_revisions ORDER BY revision DESC LIMIT 1`))
}

func (s *PostgresStore) Commit(ctx context.Context, expectedRevision uint64, payload json.RawMessage, createdBy string) (Revision, error) {
	if !json.Valid(payload) {
		return Revision{}, errors.New("configuration payload is not valid JSON")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Revision{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, revisionLockID); err != nil {
		return Revision{}, err
	}

	var current uint64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision), 0) FROM config_revisions`).Scan(&current)
	if err != nil {
		return Revision{}, err
	}
	if current != expectedRevision {
		return Revision{}, ErrRevisionConflict
	}

	hash := sha256.Sum256(payload)
	next := current + 1
	revision, err := scanRevision(tx.QueryRowContext(ctx, `
		INSERT INTO config_revisions (revision, payload, payload_sha256, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING revision, payload, payload_sha256, created_at, created_by`,
		next, []byte(payload), hex.EncodeToString(hash[:]), createdBy))
	if err != nil {
		return Revision{}, err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_notify('zoraxy_config_revision', $1)`, strconv.FormatUint(revision.ID, 10)); err != nil {
		return Revision{}, err
	}
	if err := tx.Commit(); err != nil {
		return Revision{}, err
	}
	return revision, nil
}

// Watch polls the immutable revision table. PostgreSQL NOTIFY wakes dedicated
// drivers faster, while polling guarantees recovery after disconnects and does
// not tie this package to a specific SQL driver.
func (s *PostgresStore) Watch(ctx context.Context, afterRevision uint64) (<-chan Revision, <-chan error) {
	revisions := make(chan Revision, 1)
	errorsOut := make(chan error, 1)
	go func() {
		defer close(revisions)
		defer close(errorsOut)
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		last := afterRevision
		for {
			rows, err := s.db.QueryContext(ctx, `
				SELECT revision, payload, payload_sha256, created_at, created_by
				FROM config_revisions WHERE revision > $1 ORDER BY revision`, last)
			if err != nil {
				select {
				case errorsOut <- err:
				case <-ctx.Done():
				}
				return
			}
			for rows.Next() {
				revision, err := scanRevision(rows)
				if err != nil {
					rows.Close()
					select {
					case errorsOut <- err:
					case <-ctx.Done():
					}
					return
				}
				select {
				case revisions <- revision:
					last = revision.ID
				case <-ctx.Done():
					rows.Close()
					return
				}
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				select {
				case errorsOut <- err:
				case <-ctx.Done():
				}
				return
			}
			if err := rows.Close(); err != nil {
				select {
				case errorsOut <- err:
				case <-ctx.Done():
				}
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()
	return revisions, errorsOut
}
