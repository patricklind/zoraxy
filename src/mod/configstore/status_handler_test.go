package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type statusHandlerStore struct {
	statuses []NodeStatus
	err      error
}

func (s statusHandlerStore) UpsertNodeStatus(context.Context, NodeStatus) error { return s.err }
func (s statusHandlerStore) ListNodeStatuses(context.Context) ([]NodeStatus, error) {
	return s.statuses, s.err
}

func TestNodeStatusHandlerListsConvergenceState(t *testing.T) {
	handler := NewNodeStatusHandler(statusHandlerStore{statuses: []NodeStatus{{
		NodeID:          "8f196376-f209-4e21-ae03-fec14d77d0d7",
		NodeRole:        "data-plane",
		ConfigRevision:  7,
		AppliedRevision: 6,
		LastError:       "listener conflict",
	}}})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response nodeStatusResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Nodes) != 1 || response.Nodes[0].AppliedRevision != 6 || response.Nodes[0].LastError == "" {
		t.Fatalf("response = %+v", response)
	}
}

func TestNodeStatusHandlerReportsUnavailableStore(t *testing.T) {
	handler := NewNodeStatusHandler(statusHandlerStore{err: errors.New("database unavailable")})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
