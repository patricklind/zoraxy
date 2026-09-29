package configstore

import (
	"encoding/json"
	"net/http"
)

type nodeStatusResponse struct {
	Nodes []NodeStatus `json:"nodes"`
}

// NodeStatusHandler exposes convergence state for authenticated management
// routing. Authentication and management-network restriction stay with the
// caller that mounts the handler.
type NodeStatusHandler struct {
	store NodeStatusStore
}

func NewNodeStatusHandler(store NodeStatusStore) *NodeStatusHandler {
	return &NodeStatusHandler{store: store}
}

func (h *NodeStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	statuses, err := h.store.ListNodeStatuses(r.Context())
	if err != nil {
		http.Error(w, "node status store unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(nodeStatusResponse{Nodes: statuses})
}
