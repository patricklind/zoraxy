# Active/active control-plane migration

This directory is the implementation boundary for phase 2. It does not turn
the legacy file-backed process into an active/active system merely by starting
another container.

`src/mod/configstore` defines immutable revisions, compare-and-swap commits,
PostgreSQL transactions, change watching and failure-safe activation. The SQL
migration defines the authoritative configuration, per-node convergence state,
encrypted certificate revisions and a single certificate-controller lease.

The repository now persists desired/applied revision and activation errors for
each data node. `NodeStatusHandler` provides the authenticated management API
contract for listing that state; it must be mounted by the control-plane
bootstrap when that process is introduced.

`configstore.ControlPlane` now supplies that bootstrap boundary. Mount it on
Zoraxy's authenticated management router to expose:

- `GET`/`PUT /api/cluster/config` for the current immutable revision and
  compare-and-swap commits. `PUT` requires an exact `If-Match` revision.
- `GET /api/cluster/nodes` for desired/applied revision and activation errors
  from every reporting data node.

`configstore.AtomicActivator` enforces the activation lifecycle: build and
validate a complete candidate off-path, atomically swap through a runtime
adapter, discard rejected candidates and retire the old runtime only after a
successful swap. Zoraxy's routing adapter now publishes the root route and all
host routes as one atomic snapshot without restarting the TCP/UDP listeners.
Requests already in progress retain their previous snapshot.

Routing revision payload version 1 has this top-level shape:

```json
{
  "version": 1,
  "root": { "ProxyType": 0, "RootOrMatchingDomain": "/" },
  "hosts": [
    { "ProxyType": 1, "RootOrMatchingDomain": "app.example.com" }
  ]
}
```

`ProxyType` uses Zoraxy's existing JSON representation: `0` is the root route
and `1` is a host route.

Unknown top-level fields, unsupported versions, invalid proxy types, empty
domains and duplicate case-insensitive host names are rejected before swap.
The adapter remains disabled in the main process until PostgreSQL bootstrap,
authentication and the migration mode are configured explicitly.

Run the PostgreSQL-backed repository and convergence tests entirely in Docker:

```sh
docker compose -f deploy/active-active/compose.test.yaml up \
  --build --abort-on-container-exit --exit-code-from configstore-test
docker compose -f deploy/active-active/compose.test.yaml down --volumes
```

## Required implementation order

1. Inventory every write currently made to `sys.db` and `conf/`. Move one
   complete domain at a time behind repositories; never dual-write silently.
2. Serialize all routing, access, authentication and TLS inputs into one
   versioned configuration document and validate it before `Commit`.
3. Require `If-Match: <revision>` on control-plane mutations. Map
   `configstore.ErrRevisionConflict` to HTTP 409 and return the current
   revision in the response.
4. Build a complete candidate router off-path on each data node. Call the
   `Activator` only after parsing, certificate decryption and listener conflict
   checks pass; swap the runtime atomically.
5. Update `node_status` after activation. A rejected revision records
   `last_error` and keeps the last applied router serving.
6. Move ACME to one lease-elected certificate-controller. Store only encrypted
   private keys; obtain the encryption key from an external secret provider,
   never from PostgreSQL or the image.
7. Export logs, statistics and uptime data to external sinks. They are not part
   of the configuration transaction or its RPO guarantee.
8. Put both data nodes behind a redundant L4 frontend supporting TCP and UDP.
   Test QUIC and every configured stream-proxy port explicitly.

## Migration safety

Start with PostgreSQL in shadow/read-only comparison mode. Compare its rendered
configuration hash with the legacy file configuration on every change. Cut
management writes to the control plane only after the hashes remain identical
through a full test cycle. Keep a reversible export to the legacy ZIP format
until the active/active acceptance suite has passed.

Do not enable active/active production traffic until PostgreSQL synchronous
replication, DCS quorum, secrets, certificate-controller election and L4
failover have each been tested independently.
