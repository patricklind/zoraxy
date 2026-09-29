package configstore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type controlPlaneStore struct {
	handlerStore
	statuses []NodeStatus
}

type testManagementRouter struct {
	mux *http.ServeMux
}

func (r testManagementRouter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) error {
	r.mux.HandleFunc(pattern, handler)
	return nil
}

func (s *controlPlaneStore) UpsertNodeStatus(_ context.Context, status NodeStatus) error {
	s.statuses = append(s.statuses, status)
	return nil
}

func (s *controlPlaneStore) ListNodeStatuses(context.Context) ([]NodeStatus, error) {
	return append([]NodeStatus(nil), s.statuses...), nil
}

func TestControlPlaneRegistersRevisionAndNodeStatusAPIs(t *testing.T) {
	store := &controlPlaneStore{
		handlerStore: handlerStore{current: Revision{ID: 4}},
		statuses:     []NodeStatus{{NodeID: "node-a", NodeRole: "data-plane", ConfigRevision: 4, AppliedRevision: 4}},
	}
	controlPlane, err := NewControlPlane(store, store, nil, func(*http.Request) string { return "admin" })
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	if err := controlPlane.RegisterManagementAPI(testManagementRouter{mux: mux}); err != nil {
		t.Fatal(err)
	}

	current := httptest.NewRecorder()
	mux.ServeHTTP(current, httptest.NewRequest(http.MethodGet, RevisionAPIPath, nil))
	if current.Code != http.StatusOK || current.Header().Get("ETag") != `"4"` {
		t.Fatalf("revision response status=%d etag=%q body=%s", current.Code, current.Header().Get("ETag"), current.Body.String())
	}

	commitRequest := httptest.NewRequest(http.MethodPut, RevisionAPIPath, strings.NewReader(`{"routes":[]}`))
	commitRequest.Header.Set("If-Match", `"4"`)
	committed := httptest.NewRecorder()
	mux.ServeHTTP(committed, commitRequest)
	if committed.Code != http.StatusCreated || committed.Header().Get("ETag") != `"5"` {
		t.Fatalf("commit response status=%d etag=%q body=%s", committed.Code, committed.Header().Get("ETag"), committed.Body.String())
	}

	nodes := httptest.NewRecorder()
	mux.ServeHTTP(nodes, httptest.NewRequest(http.MethodGet, NodeStatusAPIPath, nil))
	if nodes.Code != http.StatusOK {
		t.Fatalf("nodes response status=%d body=%s", nodes.Code, nodes.Body.String())
	}
	var response nodeStatusResponse
	if err := json.NewDecoder(nodes.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Nodes) != 1 || response.Nodes[0].AppliedRevision != 4 {
		t.Fatalf("nodes response = %+v", response)
	}
}

func TestControlPlaneRequiresRuntimeActivator(t *testing.T) {
	store := &controlPlaneStore{}
	controlPlane, err := NewControlPlane(store, store, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := controlPlane.RunDataNode(context.Background(), nil, NodeStatus{}); err == nil {
		t.Fatal("RunDataNode accepted a nil activator")
	}
}
