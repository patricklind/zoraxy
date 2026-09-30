package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNodeRoleFallsBackToConfigstoreMode(t *testing.T) {
	t.Setenv("ZORAXY_NODE_ROLE", "")
	t.Setenv("ZORAXY_CONFIGSTORE_MODE", "certificate-controller")
	if got := nodeRole(); got != "certificate-controller" {
		t.Fatalf("node role = %q", got)
	}
}

func TestLivenessReportsCertificateControllerLeadership(t *testing.T) {
	certificateControllerLeader.Store(true)
	t.Cleanup(func() { certificateControllerLeader.Store(false) })
	recorder := httptest.NewRecorder()
	handleLiveness(recorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	var response healthResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if !response.CertificateControllerLeader {
		t.Fatal("controller leadership was not reported")
	}
}

func TestLivenessIsPublicAndDoesNotDependOnReadiness(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handleLiveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "live" {
		t.Fatalf("health status = %q, want live", response.Status)
	}
}

func TestReadinessReportsFailedLocalChecks(t *testing.T) {
	handler := readinessHandler(func() map[string]bool {
		return map[string]bool{"database": true, "proxy_listener": false}
	})
	rec := httptest.NewRecorder()

	handler(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestReadinessSucceedsWhenAllLocalChecksPass(t *testing.T) {
	handler := readinessHandler(func() map[string]bool {
		return map[string]bool{
			"database":       true,
			"proxy_config":   true,
			"proxy_listener": true,
		}
	})
	rec := httptest.NewRecorder()

	handler(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHealthEndpointsRejectMutationMethods(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		"live":  handleLiveness,
		"ready": readinessHandler(func() map[string]bool { return map[string]bool{"ok": true} }),
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler(rec, httptest.NewRequest(http.MethodPost, "/health/"+name, nil))
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestClusterStatusExposesConvergenceFields(t *testing.T) {
	desiredConfigRevision.Store(9)
	appliedConfigRevision.Store(7)
	t.Cleanup(func() {
		desiredConfigRevision.Store(0)
		appliedConfigRevision.Store(0)
	})

	rec := httptest.NewRecorder()
	handleClusterStatus(rec, httptest.NewRequest(http.MethodGet, "/api/cluster/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.ConfigRevision != 9 || response.AppliedRevision != 7 {
		t.Fatalf("revisions = %d/%d, want 9/7", response.ConfigRevision, response.AppliedRevision)
	}
}
