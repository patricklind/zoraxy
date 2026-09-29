package main

import (
	"context"
	"encoding/json"
	"testing"

	"imuslab.com/zoraxy/mod/configstore"
	"imuslab.com/zoraxy/mod/dynamicproxy"
)

func routingTestRouter(t *testing.T) *dynamicproxy.Router {
	t.Helper()
	router, err := dynamicproxy.NewDynamicProxy(dynamicproxy.RouterOption{})
	if err != nil {
		t.Fatal(err)
	}
	return router
}

func routingRevisionPayload(t *testing.T, host string) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(routingRevisionDocument{
		Version: routingRevisionVersion,
		Root: dynamicproxy.ProxyEndpoint{
			ProxyType:            dynamicproxy.ProxyTypeRoot,
			RootOrMatchingDomain: "/",
		},
		Hosts: []dynamicproxy.ProxyEndpoint{{
			ProxyType:            dynamicproxy.ProxyTypeHost,
			RootOrMatchingDomain: host,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestZoraxyRoutingActivatorPublishesCompleteRevision(t *testing.T) {
	router := routingTestRouter(t)
	activator, err := newZoraxyRoutingActivator(router)
	if err != nil {
		t.Fatal(err)
	}
	if err := activator.Activate(context.Background(), configstore.Revision{
		ID: 1, Payload: routingRevisionPayload(t, "APP.EXAMPLE"),
	}); err != nil {
		t.Fatal(err)
	}
	if router.RootEndpoint() == nil {
		t.Fatal("root route was not activated")
	}
	if endpoint, ok := router.LoadProxyEndpoint("app.example"); !ok || endpoint.RootOrMatchingDomain != "APP.EXAMPLE" {
		t.Fatalf("activated endpoint = %+v, %v", endpoint, ok)
	}
}

func TestZoraxyRoutingActivatorRejectsDuplicateHostWithoutChangingRuntime(t *testing.T) {
	router := routingTestRouter(t)
	activator, err := newZoraxyRoutingActivator(router)
	if err != nil {
		t.Fatal(err)
	}
	if err := activator.Activate(context.Background(), configstore.Revision{
		ID: 1, Payload: routingRevisionPayload(t, "stable.example"),
	}); err != nil {
		t.Fatal(err)
	}
	stableRoot := router.RootEndpoint()

	payload, err := json.Marshal(routingRevisionDocument{
		Version: routingRevisionVersion,
		Root: dynamicproxy.ProxyEndpoint{
			ProxyType:            dynamicproxy.ProxyTypeRoot,
			RootOrMatchingDomain: "/",
		},
		Hosts: []dynamicproxy.ProxyEndpoint{
			{ProxyType: dynamicproxy.ProxyTypeHost, RootOrMatchingDomain: "duplicate.example"},
			{ProxyType: dynamicproxy.ProxyTypeHost, RootOrMatchingDomain: "DUPLICATE.EXAMPLE"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := activator.Activate(context.Background(), configstore.Revision{ID: 2, Payload: payload}); err == nil {
		t.Fatal("duplicate host revision was accepted")
	}
	if router.RootEndpoint() != stableRoot {
		t.Fatal("rejected revision changed the active root")
	}
	if _, ok := router.LoadProxyEndpoint("stable.example"); !ok {
		t.Fatal("rejected revision removed the active host")
	}
}

func TestZoraxyRoutingBuilderRejectsUnknownFields(t *testing.T) {
	builder := zoraxyRoutingBuilder{router: routingTestRouter(t)}
	_, err := builder.Build(context.Background(), configstore.Revision{
		ID:      3,
		Payload: json.RawMessage(`{"version":1,"root":{"ProxyType":"root"},"hosts":[],"unexpected":true}`),
	})
	if err == nil {
		t.Fatal("routing document with unknown field was accepted")
	}
}
