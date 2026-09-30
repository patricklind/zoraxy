# Proxmox LXC deployment

The community helper installs the released Zoraxy binary in `/opt/zoraxy`,
runs it as `zoraxy.service`, and keeps `/opt/zoraxy` as its working directory.
The cluster bootstrap reads environment variables directly, so it works without
Docker-specific entrypoint translation.

Create the LXC from the Proxmox host with the upstream helper:

```sh
bash -c "$(curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/ct/zoraxy.sh)"
```

The helper downloads the latest published GitHub release. Verify that the
installed binary version contains the active/active bootstrap before enabling
cluster mode; unreleased repository changes are not installed by that command.
The upstream helper currently deploys the binary at `/opt/zoraxy/zoraxy`, uses
`zoraxy.service`, and its update action replaces that binary before restarting
the service. Re-check the upstream script before relying on those paths in
automation.

Inside the LXC, install the supplied systemd drop-in and environment template:

```sh
install -d -m 0750 /etc/zoraxy
install -m 0640 cluster.env.example /etc/zoraxy/cluster.env
install -d -m 0755 /etc/systemd/system/zoraxy.service.d
install -m 0644 20-cluster.conf /etc/systemd/system/zoraxy.service.d/20-cluster.conf
```

Write the complete PostgreSQL DSN as one line in
`/etc/zoraxy/configstore.dsn`, owned by root with mode `0600`. Create one
32-byte certificate encryption key, store it raw or base64-encoded in
`/etc/zoraxy/certificate.key` with the same ownership and mode, and install the
identical key on the control plane and every data node. Losing or rotating this
key without re-encrypting the stored revisions makes the private keys
unreadable. Edit
`/etc/zoraxy/cluster.env` for `control-plane`, `data-plane` or
`certificate-controller`, then run:

```sh
systemctl daemon-reload
systemctl restart zoraxy
systemctl status zoraxy --no-pager
systemctl cat zoraxy
curl --fail http://127.0.0.1:8000/health/ready
```

For a data node, readiness must report equal non-zero desired/applied revisions
and `checks.config_store=true`. The process performs PostgreSQL/schema preflight and
initial revision validation before opening proxy listeners.

The community update action replaces `/opt/zoraxy/zoraxy` and restarts the
service. The `/etc/systemd/system/zoraxy.service.d` override and `/etc/zoraxy`
secrets survive that update. Before updating, confirm the target release still
supports the configured schema version and keep the previous binary available
for rollback.

Do not configure two LXC data nodes against a shared `/opt/zoraxy` filesystem.
Each node keeps separate local runtime state; PostgreSQL is shared only for the
transactional configuration domains.

For controller candidates, keep `/opt/zoraxy/conf/certs` private because the
legacy ACME engine stages decrypted key material there during renewal. Only the
lease holder renews; a lost lease cancels the CA request. The ACME account is
still stored in that candidate's local database, so preserve it across updates
until account state has also been migrated centrally.
