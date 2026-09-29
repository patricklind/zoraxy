package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"imuslab.com/zoraxy/mod/configstore"
	"imuslab.com/zoraxy/mod/dynamicproxy"
)

const (
	routingExportAPIPath = "/api/cluster/routing/export"
	routingShadowAPIPath = "/api/cluster/routing/shadow"
)

type routingMigrationHandler struct {
	router *dynamicproxy.Router
	store  configstore.Store
}

type routingShadowResponse struct {
	DesiredRevision uint64 `json:"desired_revision"`
	RuntimeSHA256   string `json:"runtime_sha256"`
	DesiredSHA256   string `json:"desired_sha256"`
	StoredSHA256    string `json:"stored_sha256"`
	Matches         bool   `json:"matches"`
}

func canonicalPayloadSHA256(payload json.RawMessage) (string, error) {
	canonical, err := canonicalRoutingRevision(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func (h routingMigrationHandler) export(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, err := routingRevisionFromRouter(h.router)
	if err != nil {
		http.Error(w, "runtime routing configuration unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(append(payload, '\n'))
}

func (h routingMigrationHandler) shadow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	runtimePayload, err := routingRevisionFromRouter(h.router)
	if err != nil {
		http.Error(w, "runtime routing configuration unavailable", http.StatusServiceUnavailable)
		return
	}
	revision, err := h.store.Current(r.Context())
	if errors.Is(err, configstore.ErrNoRevision) {
		http.Error(w, "no configuration revision exists", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "configuration store unavailable", http.StatusServiceUnavailable)
		return
	}
	runtimeHash, err := canonicalPayloadSHA256(runtimePayload)
	if err != nil {
		http.Error(w, "runtime routing configuration is invalid", http.StatusInternalServerError)
		return
	}
	desiredHash, err := canonicalPayloadSHA256(revision.Payload)
	if err != nil {
		http.Error(w, "desired routing configuration is invalid", http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(routingShadowResponse{
		DesiredRevision: revision.ID,
		RuntimeSHA256:   runtimeHash,
		DesiredSHA256:   desiredHash,
		StoredSHA256:    revision.SHA256,
		Matches:         runtimeHash == desiredHash,
	})
}
