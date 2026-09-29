package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"imuslab.com/zoraxy/mod/configstore"
	"imuslab.com/zoraxy/mod/dynamicproxy"
)

const routingRevisionVersion = 1

type routingRevisionDocument struct {
	Version int                          `json:"version"`
	Root    dynamicproxy.ProxyEndpoint   `json:"root"`
	Hosts   []dynamicproxy.ProxyEndpoint `json:"hosts"`
}

type zoraxyRoutingCandidate struct {
	snapshot dynamicproxy.RoutingSnapshot
}

func (c *zoraxyRoutingCandidate) Close() error {
	// ProxyEndpoint transports do not currently expose lifecycle cleanup. They
	// become collectible after in-flight requests release the old snapshot.
	return nil
}

type zoraxyRoutingBuilder struct {
	router *dynamicproxy.Router
}

func decodeRoutingRevisionDocument(payload json.RawMessage) (routingRevisionDocument, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var document routingRevisionDocument
	if err := decoder.Decode(&document); err != nil {
		return routingRevisionDocument{}, fmt.Errorf("decode routing document: %w", err)
	}
	if err := ensureJSONDocumentEnd(decoder); err != nil {
		return routingRevisionDocument{}, err
	}
	if document.Version != routingRevisionVersion {
		return routingRevisionDocument{}, fmt.Errorf("unsupported routing document version %d", document.Version)
	}
	if document.Root.ProxyType != dynamicproxy.ProxyTypeRoot {
		return routingRevisionDocument{}, errors.New("routing document root must use root proxy type")
	}
	seen := make(map[string]struct{}, len(document.Hosts))
	for index := range document.Hosts {
		host := &document.Hosts[index]
		if host.ProxyType != dynamicproxy.ProxyTypeHost {
			return routingRevisionDocument{}, fmt.Errorf("host %d must use host proxy type", index)
		}
		key := strings.ToLower(strings.TrimSpace(host.RootOrMatchingDomain))
		if key == "" {
			return routingRevisionDocument{}, fmt.Errorf("host %d has an empty matching domain", index)
		}
		if _, exists := seen[key]; exists {
			return routingRevisionDocument{}, fmt.Errorf("duplicate host matching domain %q", key)
		}
		seen[key] = struct{}{}
	}
	return document, nil
}

func sortRoutingHosts(hosts []dynamicproxy.ProxyEndpoint) {
	sort.Slice(hosts, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(hosts[i].RootOrMatchingDomain))
		right := strings.ToLower(strings.TrimSpace(hosts[j].RootOrMatchingDomain))
		return left < right
	})
}

func canonicalRoutingRevision(payload json.RawMessage) (json.RawMessage, error) {
	document, err := decodeRoutingRevisionDocument(payload)
	if err != nil {
		return nil, err
	}
	sortRoutingHosts(document.Hosts)
	canonical, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("encode canonical routing document: %w", err)
	}
	return canonical, nil
}

func routingRevisionFromRouter(router *dynamicproxy.Router) (json.RawMessage, error) {
	if router == nil {
		return nil, errors.New("dynamic proxy router is required")
	}
	snapshot := router.CurrentRoutingSnapshot()
	if snapshot.Root == nil {
		return nil, errors.New("dynamic proxy root route is required")
	}
	document := routingRevisionDocument{
		Version: routingRevisionVersion,
		Root:    *dynamicproxy.CopyEndpoint(snapshot.Root),
		Hosts:   []dynamicproxy.ProxyEndpoint{},
	}
	for _, endpoint := range snapshot.Endpoints {
		if endpoint != nil {
			document.Hosts = append(document.Hosts, *dynamicproxy.CopyEndpoint(endpoint))
		}
	}
	sortRoutingHosts(document.Hosts)
	payload, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("encode routing revision: %w", err)
	}
	return payload, nil
}

func (b zoraxyRoutingBuilder) Build(_ context.Context, revision configstore.Revision) (configstore.Candidate, error) {
	document, err := decodeRoutingRevisionDocument(revision.Payload)
	if err != nil {
		return nil, err
	}

	root, err := b.router.PrepareProxyRoute(&document.Root)
	if err != nil {
		return nil, fmt.Errorf("prepare root route: %w", err)
	}
	snapshot := dynamicproxy.RoutingSnapshot{
		Root:      root,
		Endpoints: make(map[string]*dynamicproxy.ProxyEndpoint, len(document.Hosts)),
	}
	for index := range document.Hosts {
		host := &document.Hosts[index]
		key := strings.ToLower(strings.TrimSpace(host.RootOrMatchingDomain))
		prepared, err := b.router.PrepareProxyRoute(host)
		if err != nil {
			return nil, fmt.Errorf("prepare host %q: %w", key, err)
		}
		snapshot.Endpoints[key] = prepared
	}
	return &zoraxyRoutingCandidate{snapshot: snapshot}, nil
}

func ensureJSONDocumentEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); errors.Is(err, io.EOF) {
		return nil
	} else if err != nil {
		return fmt.Errorf("decode trailing routing data: %w", err)
	}
	return errors.New("routing document contains multiple JSON values")
}

type zoraxyRoutingRuntime struct {
	router *dynamicproxy.Router
}

func (r zoraxyRoutingRuntime) Swap(_ context.Context, candidate configstore.Candidate) (configstore.Candidate, error) {
	routingCandidate, ok := candidate.(*zoraxyRoutingCandidate)
	if !ok {
		return nil, errors.New("unsupported routing candidate type")
	}
	previous, err := r.router.SwapRoutingSnapshot(routingCandidate.snapshot)
	if err != nil {
		return nil, err
	}
	return &zoraxyRoutingCandidate{snapshot: previous}, nil
}

func newZoraxyRoutingActivator(router *dynamicproxy.Router) (*configstore.AtomicActivator, error) {
	if router == nil {
		return nil, errors.New("dynamic proxy router is required")
	}
	return configstore.NewAtomicActivator(
		zoraxyRoutingBuilder{router: router},
		zoraxyRoutingRuntime{router: router},
		func(err error) {
			if SystemWideLogger != nil {
				SystemWideLogger.PrintAndLog("config-activation", "Unable to retire previous routing revision", err)
			}
		},
	)
}
