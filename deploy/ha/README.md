# Zoraxy active/passive HA

This deployment runs exactly one Zoraxy writer across two Linux Docker hosts.
DRBD protocol C provides synchronous state replication, while
Pacemaker/Corosync controls promotion, mount, container startup and the virtual
IP. A third `corosync-qnetd` vote and working STONITH are mandatory for
automatic failover.

This is the recommended HA mode while configuration domains other than HTTP
routing still use local files/BoltDB. The PostgreSQL-backed phase-2 work is
documented separately in [`../active-active`](../active-active/README.md); do
not combine DRBD ownership with `ZORAXY_CONFIGSTORE_MODE=data-plane`.

## Required topology

<!-- markdownlint-disable MD013 -->

| Item | Requirement |
| --- | --- |
| Nodes | `zoraxy-a` and `zoraxy-b`, with identical Linux, Docker and Compose versions |
| Witness | A separate host running `corosync-qnetd` |
| Storage | One unused, equally sized block device per Zoraxy node |
| Network | Management IP per node, dedicated replication IP per node and one unused service VIP on the same L2 segment |
| Fencing | Tested IPMI, hypervisor or managed-PDU STONITH resource for each node |
| Image | A released Zoraxy image pinned by tag and full `sha256` digest |

<!-- markdownlint-enable MD013 -->

Do not use NFS, SMB or a shared Docker volume for `sys.db`. Do not start the
Compose project manually after Pacemaker owns it.

## Install files

On both Zoraxy nodes:

1. Install Docker Engine/Compose, DRBD 9, Corosync, Pacemaker, `pcs` and the
   LINBIT/heartbeat resource agents using the host distribution's packages.
2. Copy `compose.yaml` to `/etc/zoraxy-ha/compose.yaml`, `.env.example` to
   `/etc/zoraxy-ha/zoraxy.env`, `systemd/zoraxy-ha.service` to
   `/etc/systemd/system/`, and `bin/zoraxy-ha-preflight` to
   `/usr/local/libexec/`. Copy `bin/zoraxy-ha-run` to the same directory.
3. Set mode `0755` on both helper scripts and run `systemctl daemon-reload`.
   Do not enable `zoraxy-ha.service`; Pacemaker starts it.
4. Pull the exact image on both nodes and verify that `docker image inspect`
   reports the digest specified in `zoraxy.env`.
5. Restrict TCP/8000 to the management CIDR in the host firewall. Permit the
   service ports, Corosync, DRBD replication and fencing traffic explicitly.

The Compose file deliberately excludes `/var/run/docker.sock`, selects BoltDB,
sets `ZORAXY_CONFIG_BACKEND=local`, disables mDNS and uses `/health/ready` as
its healthcheck. PostgreSQL cluster roles must remain disabled in this mode.

## Configure DRBD

Copy `drbd/zoraxy.res.example` to `/etc/drbd.d/zoraxy.res` on both nodes and
replace every `REPLACE_*` value. Generate a random shared secret; never commit
it. Then create DRBD metadata and perform the documented one-time initial sync
from the node containing the authoritative Zoraxy state.

Create an ext4 filesystem on `/dev/drbd0` only once. Its mountpoint is
`/srv/zoraxy`; it must contain `sys.db`, `sys.uuid`, `conf/` and `plugins/`.
Never mount it on both nodes simultaneously.

## Configure quorum and fencing

Create the two-node Corosync cluster and attach it to the separate qdevice.
Create and test one STONITH resource per Zoraxy node before adding application
resources. A fencing test must power off or reset the exact requested peer.

After `pcs quorum status` shows the qdevice and `pcs stonith status` shows
working devices, run `bin/configure-pacemaker` once with:

```sh
ZORAXY_VIP=192.0.2.50 \
ZORAXY_VIP_CIDR=24 \
ZORAXY_VIP_NIC=eth0 \
ZORAXY_FENCE_NODE_A=fence-zoraxy-a \
ZORAXY_FENCE_NODE_B=fence-zoraxy-b \
./bin/configure-pacemaker
```

The resulting order is DRBD promotion, filesystem, systemd/Compose service,
then VIP. The reverse order is used during shutdown, so new traffic stops
before the database is unmounted.

## Backup and restore

Use the authenticated `/api/conf/export?includeDB=true` endpoint or the Web UI
on the active node. The export now streams a BoltDB read-transaction snapshot;
it does not copy the live database file. Store backups off-cluster with
encryption, retention and restore-test metadata.

Restore only while the Pacemaker resource group is stopped. Restore into an
isolated path first, validate the ZIP structure and database, replace the DRBD
state, then let Pacemaker start the stack. DRBD is replication, not backup.

## Acceptance run

Record timestamps and results for each scenario in `FAILOVER-ACCEPTANCE.md`:

1. Move `zoraxy-stack` cleanly to the peer.
2. Stop the active container and then Docker itself.
3. Power off the active host through its fencing path.
4. Disconnect only the Corosync/replication network and verify that the
   non-quorate node does not mount or serve traffic.
5. Apply a proxy and certificate change immediately before fencing the active
   host; verify the change after promotion.
6. Exercise HTTP, HTTPS, WebSocket, TCP stream proxy and UDP/QUIC.
7. Run `ZORAXY_VIP=<vip> bin/smoke-test`; readiness must return within 60s.

The ready response must return HTTP 200 with `node_role=active-passive`, equal
zero configuration revision fields, and true `database`, `proxy_config` and
`proxy_listener` values under `checks`. Revision zero is expected because this
mode uses the local backend.

An acceptance run fails if both nodes are Primary/mounted, fencing is skipped,
the VIP exists on both nodes, readiness exceeds 60 seconds, or a committed
configuration/certificate change is missing.
