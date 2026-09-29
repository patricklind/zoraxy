package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type testStore struct{ revisions chan Revision }

func (s *testStore) Current(context.Context) (Revision, error) { return Revision{}, ErrNoRevision }
func (s *testStore) Commit(context.Context, uint64, json.RawMessage, string) (Revision, error) {
	return Revision{}, errors.New("not implemented")
}
func (s *testStore) Watch(context.Context, uint64) (<-chan Revision, <-chan error) {
	errs := make(chan error)
	close(errs)
	return s.revisions, errs
}

type testActivator struct{ reject uint64 }

func (a testActivator) Activate(_ context.Context, revision Revision) error {
	if revision.ID == a.reject {
		return errors.New("invalid configuration")
	}
	return nil
}

type testNodeStatusStore struct {
	updates []NodeStatus
	err     error
}

func (s *testNodeStatusStore) UpsertNodeStatus(_ context.Context, status NodeStatus) error {
	if s.err != nil {
		return s.err
	}
	s.updates = append(s.updates, status)
	return nil
}

func (s *testNodeStatusStore) ListNodeStatuses(context.Context) ([]NodeStatus, error) {
	return append([]NodeStatus(nil), s.updates...), s.err
}

func TestFollowKeepsServingAfterRejectedRevision(t *testing.T) {
	revisions := make(chan Revision, 3)
	revisions <- Revision{ID: 1}
	revisions <- Revision{ID: 2}
	revisions <- Revision{ID: 3}
	close(revisions)

	var applied []uint64
	var rejected []uint64
	err := Follow(context.Background(), &testStore{revisions: revisions}, testActivator{reject: 2}, 0,
		func(revision Revision) { applied = append(applied, revision.ID) },
		func(revision Revision, _ error) { rejected = append(rejected, revision.ID) })
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 || applied[0] != 1 || applied[1] != 3 {
		t.Fatalf("applied revisions = %v, want [1 3]", applied)
	}
	if len(rejected) != 1 || rejected[0] != 2 {
		t.Fatalf("rejected revisions = %v, want [2]", rejected)
	}
}

func TestPostgresStoreDefaultPollInterval(t *testing.T) {
	store := NewPostgresStore(nil, 0)
	if store.pollInterval != time.Second {
		t.Fatalf("poll interval = %s, want 1s", store.pollInterval)
	}
}

func TestFollowNodeReportsDesiredRejectedAndAppliedRevisions(t *testing.T) {
	revisions := make(chan Revision, 3)
	revisions <- Revision{ID: 1}
	revisions <- Revision{ID: 2}
	revisions <- Revision{ID: 3}
	close(revisions)

	statuses := &testNodeStatusStore{}
	err := FollowNode(context.Background(), &testStore{revisions: revisions}, testActivator{reject: 2}, statuses, NodeStatus{
		NodeID:   "8f196376-f209-4e21-ae03-fec14d77d0d7",
		NodeRole: "data-plane",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(statuses.updates) != 7 {
		t.Fatalf("status updates = %d, want 7", len(statuses.updates))
	}
	rejected := statuses.updates[4]
	if rejected.ConfigRevision != 2 || rejected.AppliedRevision != 1 || rejected.LastError != "invalid configuration" {
		t.Fatalf("rejected status = %+v", rejected)
	}
	final := statuses.updates[6]
	if final.ConfigRevision != 3 || final.AppliedRevision != 3 || final.LastError != "" {
		t.Fatalf("final status = %+v", final)
	}
}

func TestFollowNodeStopsWhenConvergenceCannotBeReported(t *testing.T) {
	reportErr := errors.New("status database unavailable")
	err := FollowNode(context.Background(), &testStore{revisions: make(chan Revision)}, testActivator{},
		&testNodeStatusStore{err: reportErr}, NodeStatus{NodeID: "node-1", NodeRole: "data-plane"})
	if !errors.Is(err, reportErr) {
		t.Fatalf("error = %v, want %v", err, reportErr)
	}
}
