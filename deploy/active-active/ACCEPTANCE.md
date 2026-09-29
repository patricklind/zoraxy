# Active/active acceptance record

This record covers only capabilities already migrated to PostgreSQL. Mark all
unmigrated configuration domains as blocked rather than treating HTTP-routing
success as full product acceptance.

| Field | Value |
| --- | --- |
| Date and operator | |
| Zoraxy version and image digest | |
| PostgreSQL version/topology | |
| DCS/witness topology | |
| Control-plane node | |
| Data-node IDs | |
| L4 frontend version/topology | |
| Initial revision and SHA-256 | |

| Test | Expected result | Evidence | Pass/fail |
| --- | --- | --- | --- |
| Schema verification | Exact supported schema version accepted | | |
| Missing/incorrect schema | Process exits before proxy listener opens | | |
| Missing initial revision | Data node exits before proxy listener opens | | |
| Valid revision N | Both nodes report desired/applied N | | |
| Valid revision N+1 | Both nodes converge without listener restart | | |
| Invalid revision N+2 | N+1 keeps serving; node error is recorded | | |
| Stale `If-Match` | HTTP 409; no new revision | | |
| Missing `If-Match` | HTTP 428; no new revision | | |
| Unauthenticated write | Rejected by management authentication | | |
| Missing/invalid CSRF token | Write rejected | | |
| Data-node restart | Current revision loads before listener opens | | |
| PostgreSQL primary failure | Synchronous replica promotes; nodes recover | | |
| One data-node failure | L4 removes only failed node | | |
| Long-lived request during update | Existing request completes on old snapshot | | |
| HTTP/HTTPS/WebSocket | Traffic succeeds through both nodes | | |
| Rollback revision | New revision restores last known-good behavior | | |
| Full Go race suite | `go test -race ./...` passes in Docker | | |
| Static/build gates | `go vet ./...` and `go build ./...` pass in Docker | | |

Full active/active acceptance remains blocked until TLS/certificates, ACME,
access rules, authentication, redirects and stream proxies use the same
transactional revision lifecycle.
