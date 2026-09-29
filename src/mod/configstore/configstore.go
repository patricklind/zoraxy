package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNoRevision       = errors.New("no configuration revision exists")
	ErrRevisionConflict = errors.New("configuration revision conflict")
)

type Revision struct {
	ID        uint64          `json:"id"`
	Payload   json.RawMessage `json:"payload"`
	SHA256    string          `json:"sha256"`
	CreatedAt time.Time       `json:"created_at"`
	CreatedBy string          `json:"created_by"`
}

// Store is the authoritative control-plane contract. Commit must atomically
// compare expectedRevision and create the next immutable revision.
type Store interface {
	Current(ctx context.Context) (Revision, error)
	Commit(ctx context.Context, expectedRevision uint64, payload json.RawMessage, createdBy string) (Revision, error)
	Watch(ctx context.Context, afterRevision uint64) (<-chan Revision, <-chan error)
}

// Activator validates and atomically swaps a complete runtime configuration.
// A failed activation must leave the previous revision serving traffic.
type Activator interface {
	Activate(ctx context.Context, revision Revision) error
}

type ActivatorFunc func(context.Context, Revision) error

func (f ActivatorFunc) Activate(ctx context.Context, revision Revision) error {
	return f(ctx, revision)
}

// NodeStatusStore records data-plane convergence independently from revision
// commits. Reporting failures must never roll back an already active runtime.
type NodeStatusStore interface {
	UpsertNodeStatus(ctx context.Context, status NodeStatus) error
	ListNodeStatuses(ctx context.Context) ([]NodeStatus, error)
}

type NodeStatus struct {
	NodeID          string    `json:"node_id"`
	NodeRole        string    `json:"node_role"`
	ConfigRevision  uint64    `json:"config_revision"`
	AppliedRevision uint64    `json:"applied_revision"`
	LastError       string    `json:"last_error,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}
