package configstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresStoreIntegration(t *testing.T) {
	dsn := os.Getenv("CONFIGSTORE_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("CONFIGSTORE_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE config_revisions, node_status`); err != nil {
		t.Fatal(err)
	}

	store := NewPostgresStore(db, 10*time.Millisecond)
	revision, err := store.Commit(ctx, 0, json.RawMessage(`{"routes":[]}`), "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	if revision.ID != 1 || len(revision.SHA256) != 64 {
		t.Fatalf("revision = %+v", revision)
	}
	if _, err := store.Commit(ctx, 0, json.RawMessage(`{"routes":[]}`), "stale-writer"); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale commit error = %v, want %v", err, ErrRevisionConflict)
	}

	watched, failures := store.Watch(ctx, 0)
	select {
	case got := <-watched:
		if got.ID != revision.ID {
			t.Fatalf("watched revision = %d, want %d", got.ID, revision.ID)
		}
	case err := <-failures:
		t.Fatalf("watch failed: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	status := NodeStatus{
		NodeID:          "8f196376-f209-4e21-ae03-fec14d77d0d7",
		NodeRole:        "data-plane",
		ConfigRevision:  1,
		AppliedRevision: 1,
	}
	if err := store.UpsertNodeStatus(ctx, status); err != nil {
		t.Fatal(err)
	}
	statuses, err := store.ListNodeStatuses(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].NodeID != status.NodeID || statuses[0].AppliedRevision != 1 {
		t.Fatalf("statuses = %+v", statuses)
	}
}
