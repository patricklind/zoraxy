package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (b zoraxyRoutingBuilder) Build(_ context.Context, revision configstore.Revision) (configstore.Candidate, error) {
	decoder := json.NewDecoder(bytes.NewReader(revision.Payload))
	decoder.DisallowUnknownFields()
	var document routingRevisionDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode routing document: %w", err)
	}
	if err := ensureJSONDocumentEnd(decoder); err != nil {
		return nil, err
	}
	if document.Version != routingRevisionVersion {
		return nil, fmt.Errorf("unsupported routing document version %d", document.Version)
	}
	if document.Root.ProxyType != dynamicproxy.ProxyTypeRoot {
		return nil, errors.New("routing document root must use root proxy type")
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
		if host.ProxyType != dynamicproxy.ProxyTypeHost {
			return nil, fmt.Errorf("host %d must use host proxy type", index)
		}
		key := strings.ToLower(strings.TrimSpace(host.RootOrMatchingDomain))
		if key == "" {
			return nil, fmt.Errorf("host %d has an empty matching domain", index)
		}
		if _, exists := snapshot.Endpoints[key]; exists {
			return nil, fmt.Errorf("duplicate host matching domain %q", key)
		}
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
