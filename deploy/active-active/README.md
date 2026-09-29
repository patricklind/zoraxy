# Active/active control-plane migration

This directory contains the implemented phase-2 foundation. It can run a
separate authenticated control-plane process and PostgreSQL-backed HTTP-routing
data nodes. It does not yet make certificates, access rules, redirects, stream
proxies, users, plugins or ACME state transactional. Do not call the overall
product production-ready active/active until those remaining domains are moved.

`src/mod/configstore` defines immutable revisions, compare-and-swap commits,
PostgreSQL transactions, change watching and failure-safe activation. The SQL
migration defines the authoritative configuration, per-node convergence state,
encrypted certificate revisions and a single certificate-controller lease.

The repository persists desired/applied revision and activation errors for
each data node. The main process now mounts `NodeStatusHandler` automatically
when started in `control-plane` mode.

`configstore.ControlPlane` is mounted on Zoraxy's authenticated management
router to expose:

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
The adapter is disabled by default. Select the backend explicitly with
`ZORAXY_CONFIG_BACKEND=local|postgresql`. `local` retains the original
BoltDB/LevelDB and file-backed behavior. `postgresql` currently becomes
authoritative for the migrated HTTP-routing document; a local database remains
in use for configuration domains that have not been migrated yet.

With the PostgreSQL backend, enable exactly one role per process:

- `ZORAXY_CONFIGSTORE_MODE=control-plane` mounts the revision and node APIs on
  Zoraxy's authenticated management router. Startup is rejected when `NOAUTH`
  is enabled.
- `ZORAXY_CONFIGSTORE_MODE=data-plane` requires an existing revision, validates
  and activates it before readiness succeeds, then follows later revisions.

Both modes require `ZORAXY_CONFIGSTORE_MIGRATION_MODE=verify|apply` and a
PostgreSQL connection. Supply the full connection string through a Docker
secret with `ZORAXY_CONFIGSTORE_DSN_FILE`; `ZORAXY_CONFIGSTORE_DSN` exists for
test environments but exposes the secret through the process environment.
`verify` is the production default: it refuses startup unless schema version 1
is already installed. `apply` performs the idempotent version-1 migration in a
single transaction. A data node becomes unready if its revision follower exits.

Docker is not required on the target node. Proxmox LXC installations created
with the community helper are supported through a persistent systemd drop-in;
see [`lxc/README.md`](lxc/README.md).

## Startup guarantees

Cluster mode is fail-closed:

1. Parse role, migration mode, durations and exactly one DSN source.
2. Connect to PostgreSQL and apply or verify the exact schema version.
3. In `data-plane` mode, require an existing revision before starting the
   dynamic proxy.
4. Construct the router without loading legacy HTTP routing files.
5. Validate and atomically install the PostgreSQL revision.
6. Release the listener barrier and report readiness only after listeners are
   active.
7. Poll for later revisions. A failed candidate records `last_error`, leaves
   the prior snapshot serving and allows a later valid revision to recover.

`control-plane` mode refuses to start with `NOAUTH=true`. Its write API is also
inside the existing CSRF middleware. Keep TCP/8000 restricted to the management
network.

## Initial deployment

Apply the schema once from a controlled migration job:

```sh
docker run --rm --network <postgres-network> \
  -e ZORAXY_CONFIG_BACKEND=postgresql \
  -e ZORAXY_CONFIGSTORE_MODE=control-plane \
  -e ZORAXY_CONFIGSTORE_MIGRATION_MODE=apply \
  -e ZORAXY_CONFIGSTORE_DSN_FILE=/run/secrets/configstore-dsn \
  -v ./secrets/configstore-dsn:/run/secrets/configstore-dsn:ro \
  <pinned-zoraxy-image>
```

Stop that job after schema creation and change every long-running process to
`ZORAXY_CONFIGSTORE_MIGRATION_MODE=verify`. The first data node will refuse to
start until revision 1 exists. Create it through authenticated
`PUT /api/cluster/config` on the control plane with `If-Match: "0"`; include
the normal Zoraxy CSRF token/cookie required by the management interface.

After the first commit, start data nodes with separate local state volumes and
the same PostgreSQL DSN secret. Do not share their BoltDB/config directories.
Check each node directly before adding it to the L4 frontend:

```sh
curl --fail http://<node-management-ip>:8000/health/ready
```

The response must include `node_role=data-plane`, equal non-zero
`config_revision` and `applied_revision`, and `checks.config_store=true`.

## Environment reference

| Variable | Required value or default | Purpose |
| --- | --- | --- |
| `ZORAXY_CONFIG_BACKEND` | `local` (default) or `postgresql` | Selects the authoritative configuration backend |
| `ZORAXY_CONFIGSTORE_MODE` | `control-plane` or `data-plane` with PostgreSQL | Selects exactly one process role |
| `ZORAXY_CONFIGSTORE_MIGRATION_MODE` | `verify` or `apply` | Verifies the exact schema or installs schema version 1 |
| `ZORAXY_CONFIGSTORE_DSN_FILE` | Preferred secret-file path | Reads the complete PostgreSQL DSN without an environment secret |
| `ZORAXY_CONFIGSTORE_DSN` | Test environments only | Supplies the DSN directly; mutually exclusive with the file variable |
| `ZORAXY_CONFIGSTORE_POLL_INTERVAL` | `1s` | Positive Go duration between revision checks |
| `ZORAXY_CONFIGSTORE_STARTUP_TIMEOUT` | `10s` | Positive Go duration for PostgreSQL/schema startup preflight |
| `ZORAXY_NODE_ROLE` | `standalone` when unset | Label exposed in health and status responses |

`DB=boltdb|leveldb|auto` still controls the local database used by unmigrated
domains. `DB=postgresql` is rejected with an explicit startup error.

## API behavior

`GET /api/cluster/config` returns the authoritative revision and an `ETag`.
Send that exact revision with the next mutation:

```http
PUT /api/cluster/config HTTP/1.1
If-Match: "7"
Content-Type: application/json

{"version":1,"root":{"ProxyType":0,"RootOrMatchingDomain":"/"},"hosts":[]}
```

A missing `If-Match` returns 428, a stale revision returns 409, invalid JSON
returns 400, and a routing document that cannot be activated returns 422. A
successful commit returns 201 with the new `ETag`; data-plane convergence is
reported separately by `GET /api/cluster/nodes`.

Two additional authenticated, read-only migration endpoints are available on
the control plane:

- `GET /api/cluster/routing/export` exports the current local HTTP-routing
  runtime as a deterministically ordered version-1 revision document. Treat the
  response as secret because proxy authentication data may be present.
- `GET /api/cluster/routing/shadow` compares canonical SHA-256 values for the
  local runtime and current PostgreSQL revision. JSON whitespace and host order
  do not cause false mismatches. `matches=true` is a migration signal, not proof
  that the still-unmigrated configuration domains are equivalent.

Run the PostgreSQL-backed repository and convergence tests entirely in Docker:

```sh
docker compose -f deploy/active-active/compose.test.yaml up \
  --build --abort-on-container-exit --exit-code-from configstore-test
docker compose -f deploy/active-active/compose.test.yaml down --volumes
```

Run the complete Go safety gates from a writable copy inside Docker. Several
legacy tests create temporary databases beside their package, so a read-only
source mount alone is insufficient:

```sh
docker run --rm -v "$PWD":/workspace:ro golang:1.26 sh -c \
  'cp -a /workspace /tmp/zoraxy && cd /tmp/zoraxy/src && \
   go vet ./... && go test -race ./... && go build ./...'
```

## Troubleshooting

| Symptom | Cause to check first |
| --- | --- |
| Process exits before listeners open | Missing DSN, wrong schema version, invalid mode or no initial revision |
| Ready returns 503 with `checks.config_store=false` | Revision follower stopped or initial activation failed |
| Desired revision is above applied revision | Candidate validation failed; inspect `last_error` in `/api/cluster/nodes` |
| Control plane refuses startup | Management authentication is disabled (`NOAUTH=true`) |
| `DB=postgresql` startup error | PostgreSQL must be selected with `ZORAXY_CONFIG_BACKEND`, not the local DB selector |
| LXC works locally but ignores cluster variables | Missing systemd drop-in, unreadable DSN file or installed release predates this bootstrap |

## Remaining implementation order

1. Inventory every write currently made to `sys.db` and `conf/`. Move one
   complete domain at a time behind repositories; never dual-write silently.
2. Extend the versioned document beyond the implemented HTTP-routing snapshot
   to access, authentication, redirects, streams and TLS.
3. Migrate legacy management write endpoints to repository transactions; until
   then they can still change local state and must not be used for cluster
   configuration.
4. Add certificate decryption and listener-conflict validation to the existing
   off-path candidate builder.
5. Move ACME to one lease-elected certificate-controller. Store only encrypted
   private keys; obtain the encryption key from an external secret provider,
   never from PostgreSQL or the image.
6. Export logs, statistics and uptime data to external sinks. They are not part
   of the configuration transaction or its RPO guarantee.
7. Put both data nodes behind a redundant L4 frontend supporting TCP and UDP.
   Test QUIC and every configured stream-proxy port explicitly.

## Migration safety

Start with the control plane isolated from production listeners. Import a
manually reviewed routing snapshot as revision 1, start one data node outside
the L4 pool and compare its behavior with the legacy node. Legacy management
write endpoints are not intercepted yet; operational policy must make them
read-only during this stage. Add the second data node only after revision and
behavior checks pass.

Rollback is a new immutable revision containing the last known-good payload;
never delete or rewrite revision rows. If PostgreSQL or the new control plane
must be abandoned, remove data nodes from the L4 pool and restore traffic to
the untouched active/passive deployment. A data-plane node cannot safely be
switched back to legacy routing merely by changing its environment while it is
serving traffic.

## Acceptance evidence

Before traffic cutover, retain evidence for:

- schema `verify` succeeding against the PostgreSQL primary and synchronous
  replica after failover;
- two data nodes simultaneously showing the same non-zero applied revision;
- revision N+1 converging on both nodes without listener restart;
- an invalid N+2 recording an error while N+1 continues serving;
- node restart loading the current revision before opening proxy listeners;
- authenticated/CSRF-protected writes, stale `If-Match` rejection and audit
  identity in `created_by`;
- HTTP, HTTPS, WebSocket and long-lived requests through the L4 frontend;
- PostgreSQL/DCS, L4, ACME-controller and rolling-upgrade failure tests.

Record the run in [`ACCEPTANCE.md`](ACCEPTANCE.md).

Do not enable active/active production traffic until PostgreSQL synchronous
replication, DCS quorum, secrets, certificate-controller election and L4
failover have each been tested independently.
