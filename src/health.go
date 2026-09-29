package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

var (
	desiredConfigRevision atomic.Uint64
	appliedConfigRevision atomic.Uint64
)

type healthResponse struct {
	Status          string          `json:"status"`
	NodeID          string          `json:"node_id,omitempty"`
	NodeRole        string          `json:"node_role"`
	ConfigRevision  uint64          `json:"config_revision"`
	AppliedRevision uint64          `json:"applied_revision"`
	Checks          map[string]bool `json:"checks,omitempty"`
}

type readinessChecker func() map[string]bool

func nodeRole() string {
	role := strings.TrimSpace(os.Getenv("ZORAXY_NODE_ROLE"))
	if role == "" {
		return "standalone"
	}
	return role
}

func currentReadinessChecks() map[string]bool {
	checks := map[string]bool{
		"database":       false,
		"proxy_config":   false,
		"proxy_listener": false,
	}

	if sysdb != nil {
		checks["database"] = sysdb.TableExists("settings")
	}
	if dynamicProxyRouter != nil {
		checks["proxy_config"] = dynamicProxyRouter.Root != nil
		checks["proxy_listener"] = dynamicProxyRouter.IsReady()
	}
	return checks
}

func allChecksPass(checks map[string]bool) bool {
	if len(checks) == 0 {
		return false
	}
	for _, passed := range checks {
		if !passed {
			return false
		}
	}
	return true
}

func writeHealthJSON(w http.ResponseWriter, statusCode int, response healthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func handleLiveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeHealthJSON(w, http.StatusOK, healthResponse{
		Status:          "live",
		NodeID:          nodeUUID,
		NodeRole:        nodeRole(),
		ConfigRevision:  desiredConfigRevision.Load(),
		AppliedRevision: appliedConfigRevision.Load(),
	})
}

func readinessHandler(check readinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		checks := check()
		statusCode := http.StatusServiceUnavailable
		status := "not_ready"
		if allChecksPass(checks) {
			statusCode = http.StatusOK
			status = "ready"
		}
		writeHealthJSON(w, statusCode, healthResponse{
			Status:          status,
			NodeID:          nodeUUID,
			NodeRole:        nodeRole(),
			ConfigRevision:  desiredConfigRevision.Load(),
			AppliedRevision: appliedConfigRevision.Load(),
			Checks:          checks,
		})
	}
}

func handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	checks := currentReadinessChecks()
	status := "not_ready"
	if allChecksPass(checks) {
		status = "ready"
	}
	writeHealthJSON(w, http.StatusOK, healthResponse{
		Status:          status,
		NodeID:          nodeUUID,
		NodeRole:        nodeRole(),
		ConfigRevision:  desiredConfigRevision.Load(),
		AppliedRevision: appliedConfigRevision.Load(),
		Checks:          checks,
	})
}

func registerHealthEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/health/live", handleLiveness)
	mux.HandleFunc("/health/ready", readinessHandler(currentReadinessChecks))
}
