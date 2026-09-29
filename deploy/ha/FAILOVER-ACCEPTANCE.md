# Failover acceptance record

| Field | Value |
| --- | --- |
| Date | |
| Operator | |
| Zoraxy image digest | |
| Node A / Node B | |
| Witness | |
| VIP | |
| DRBD status before test | |
| Pacemaker status before test | |

<!-- markdownlint-disable MD013 -->

| Scenario | Started | Ready again | Downtime | RPO result | Pass/fail | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| Controlled switchover | | | | | | |
| Container failure | | | | | | |
| Docker failure | | | | | | |
| Active-host fencing | | | | | | |
| Network partition | | | | | | |
| Config write immediately before failure | | | | | | |
| Certificate write immediately before failure | | | | | | |
| HTTP/HTTPS/WebSocket | | | | n/a | | |
| TCP stream and UDP/QUIC | | | | n/a | | |
| Backup restore in isolation | | | | | | |

<!-- markdownlint-enable MD013 -->

Acceptance requires a single writer throughout, recovery within 60 seconds and
no loss of an acknowledged configuration or certificate write.
