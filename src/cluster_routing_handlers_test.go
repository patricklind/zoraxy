package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"imuslab.com/zoraxy/mod/configstore"
	"imuslab.com/zoraxy/mod/dynamicproxy"
)

type routingMigrationTestStore struct {
	current configstore.Revision
	err     error
}

func (s routingMigrationTestStore) Current(context.Context) (configstore.Revision, error) {
	return s.current, s.err
}

func (routingMigrationTestStore) Commit(context.Context, uint64, json.RawMessage, string) (configstore.Revision, error) {
	panic("unexpected commit")
}

func (routingMigrationTestStore) Watch(context.Context, uint64) (<-chan configstore.Revision, <-chan error) {
	panic("unexpected watch")
}

func routingMigrationTestRouter(t *testing.T, hosts ...string) *dynamicproxy.Router {
	t.Helper()
	router := routingTestRouter(t)
	endpoints := make(map[string]*dynamicproxy.ProxyEndpoint, len(hosts))
	for _, host := range hosts {
		endpoints[host] = &dynamicproxy.ProxyEndpoint{
			ProxyType:            dynamicproxy.ProxyTypeHost,
			RootOrMatchingDomain: host,
		}
	}
	if _, err := router.SwapRoutingSnapshot(dynamicproxy.RoutingSnapshot{
		Root: &dynamicproxy.ProxyEndpoint{
			ProxyType:            dynamicproxy.ProxyTypeRoot,
			RootOrMatchingDomain: "/",
		},
		Endpoints: endpoints,
	}); err != nil {
		t.Fatal(err)
	}
	return router
}

func TestRoutingRevisionExportIsDeterministic(t *testing.T) {
	first, err := routingRevisionFromRouter(routingMigrationTestRouter(t, "z.example", "A.example"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := routingRevisionFromRouter(routingMigrationTestRouter(t, "A.example", "z.example"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("exports differ:\n%s\n%s", first, second)
	}
	if strings.Index(string(first), "A.example") > strings.Index(string(first), "z.example") {
		t.Fatalf("hosts are not sorted: %s", first)
	}
}

func TestCanonicalRoutingHashIgnoresWhitespaceAndHostOrder(t *testing.T) {
	first := json.RawMessage(`{"version":1,"root":{"ProxyType":0,"RootOrMatchingDomain":"/"},"hosts":[{"ProxyType":1,"RootOrMatchingDomain":"z.example"},{"ProxyType":1,"RootOrMatchingDomain":"a.example"}]}`)
	second := json.RawMessage(`{
        "hosts": [
          {"RootOrMatchingDomain":"a.example", "ProxyType":1},
          {"RootOrMatchingDomain":"z.example", "ProxyType":1}
        ],
        "root":{"RootOrMatchingDomain":"/", "ProxyType":0},
        "version":1
      }`)
	firstHash, err := canonicalPayloadSHA256(first)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := canonicalPayloadSHA256(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash {
		t.Fatalf("canonical hashes differ: %s != %s", firstHash, secondHash)
	}
}

func TestRoutingMigrationHandlersExportAndCompareRuntime(t *testing.T) {
	router := routingMigrationTestRouter(t, "b.example", "a.example")
	payload, err := routingRevisionFromRouter(router)
	if err != nil {
		t.Fatal(err)
	}
	handler := routingMigrationHandler{
		router: router,
		store: routingMigrationTestStore{current: configstore.Revision{
			ID: 9, Payload: payload, SHA256: "stored-raw-hash",
		}},
	}

	exported := httptest.NewRecorder()
	handler.export(exported, httptest.NewRequest(http.MethodGet, routingExportAPIPath, nil))
	if exported.Code != http.StatusOK || exported.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("export status=%d headers=%v body=%s", exported.Code, exported.Header(), exported.Body.String())
	}
	if strings.TrimSpace(exported.Body.String()) != string(payload) {
		t.Fatalf("export body = %s, want %s", exported.Body.String(), payload)
	}

	shadow := httptest.NewRecorder()
	handler.shadow(shadow, httptest.NewRequest(http.MethodGet, routingShadowAPIPath, nil))
	if shadow.Code != http.StatusOK {
		t.Fatalf("shadow status=%d body=%s", shadow.Code, shadow.Body.String())
	}
	var response routingShadowResponse
	if err := json.NewDecoder(shadow.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if !response.Matches || response.DesiredRevision != 9 || response.RuntimeSHA256 == "" || response.RuntimeSHA256 != response.DesiredSHA256 {
		t.Fatalf("shadow response = %+v", response)
	}
}

func TestRoutingShadowReportsMissingRevision(t *testing.T) {
	handler := routingMigrationHandler{
		router: routingMigrationTestRouter(t),
		store:  routingMigrationTestStore{err: configstore.ErrNoRevision},
	}
	response := httptest.NewRecorder()
	handler.shadow(response, httptest.NewRequest(http.MethodGet, routingShadowAPIPath, nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
