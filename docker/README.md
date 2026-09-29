# Zoraxy Docker

[![Repo](https://img.shields.io/badge/Docker-Repo-007EC6?labelColor-555555&color-007EC6&logo=docker&logoColor=fff&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Version](https://img.shields.io/docker/v/zoraxydocker/zoraxy/latest?labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Size](https://img.shields.io/docker/image-size/zoraxydocker/zoraxy/latest?sort=semver&labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Pulls](https://img.shields.io/docker/pulls/zoraxydocker/zoraxy?labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)

## Usage

If you are attempting to access your service from outside your network, make sure to forward ports 80 and 443 to the Zoraxy host to allow web traffic. If you know how to do this, great! If not, find the manufacturer of your router and search on how to do that. There are too many to be listed here. Read more about it from [whatismyip](https://www.whatismyip.com/port-forwarding/).

In the examples below, make sure to update `/path/to/zoraxy/config/`. If a path is not provided, an anonymous Docker volume will be created instead to prevent data loss, but it is recommended to store the data at a defined host location or a named Docker volume.

Once setup, access the webui at `http://<host-ip>:8000` to configure Zoraxy. Change the port in the URL if you changed the management port.

### Docker Run

```
docker run -d \
  --name zoraxy \
  --restart unless-stopped \
  --add-host=host.docker.internal:host-gateway \
  -p 80:80 \
  -p 443:443 \
  -p 443:443/udp \
  -p 8000:8000 \
  -v /path/to/zoraxy/config/:/opt/zoraxy/config/ \
  -v /path/to/zoraxy/plugin/:/opt/zoraxy/plugin/ \
  -e FASTGEOIP="true" \
  -e TZ="America/New_York" \
  zoraxydocker/zoraxy:latest
```

### Docker Compose

```yml
services:
  zoraxy:
    image: zoraxydocker/zoraxy:latest
    container_name: zoraxy
    restart: unless-stopped
    ports:
      - 80:80
      - 443:443
      - 443:443/udp
      - 8000:8000
    volumes:
      - /path/to/zoraxy/config/:/opt/zoraxy/config/
      - /path/to/zoraxy/plugin/:/opt/zoraxy/plugin/
    extra_hosts:
      - "host.docker.internal:host-gateway"
    environment:
      FASTGEOIP: "true"
      TZ: "America/New_York"
```

### Ports

| Port | Details |
|:-|:-|
| `80` | HTTP traffic. |
| `443` | HTTPS traffic. |
| `443/udp` | HTTP/3 (QUIC). Only needed if HTTP/3 is enabled in Global Settings. |
| `8000` | Management interface. Can be changed with the `PORT` env. |

### Volumes

| Volume | Details |
|:-|:-|
| `/opt/zoraxy/config/` | Zoraxy configuration. |
| `/opt/zoraxy/plugin/` | Zoraxy plugins. |

### Extra Hosts
| Host | Details |
|:-|:-|
| `host.docker.internal:host-gateway` | Resolves host.docker.internal to the host’s gateway IP on the Docker bridge network, allowing containers to access services running on the host machine. |

### Docker socket

The default examples deliberately do not mount `/var/run/docker.sock`. A direct
socket mount gives the container control over the Docker host and is not needed
for normal reverse-proxy operation. If container discovery is required, place a
restricted Docker socket proxy in front of the daemon and expose only the
read-only API operations the integration uses. Do not add a writable socket
mount to either HA mode.

### Environment

Variables are the same as those in [Start Parameters](https://github.com/tobychui/zoraxy?tab=readme-ov-file#start-paramters).

| Variable | Default | Details |
|:-|:-|:-|
| `AUTORENEW` | `86400` (Integer) | ACME auto TLS/SSL certificate renew check interval. |
| `CFGUPGRADE` | `true` (Boolean) | Enable auto config upgrade if breaking change is detected. |
| `DB` | `auto` (String) | Database backend to use (leveldb, boltdb, auto) Note that fsdb will be used on unsupported platforms like RISCV (default "auto"). |
| `DOCKER` | `true` (Boolean) | Run Zoraxy in docker compatibility mode. |
| `EARLYRENEW` | `30` (Integer) | Number of days to early renew a soon expiring certificate. |
| `ENABLELOG` | `true` (Boolean) | Enable system wide logging, set to false for writing log to STDOUT only. |
| `FASTGEOIP` | `false`  (Boolean) | Enable high speed geoip lookup, requires 1GB extra memory (Not recommend for low end devices). |
| `MDNS` | `true` (Boolean) | Enable mDNS scanner and transponder. |
| `MDNSNAME` | `''` (String) | mDNS name, leave empty to use default (zoraxy_{node-uuid}.local). |
| `NOAUTH` | `false` (Boolean) | Disable authentication for management interface. |
| `PLUGIN` | `/opt/zoraxy/plugin/` (String) | Set the path for Zoraxy plugins. Only change this if you know what you are doing. |
| `PORT` | `8000` (Integer) | Management web interface listening port |
| `SSHLB` | `false` (Boolean) | Allow loopback web ssh connection (DANGER). |
| `TZ` | `Etc/UTC` (String) | Define timezone using [standard tzdata values](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). |
| `UPDATE_GEOIP` | `false` (Boolean) | Download the latest GeoIP data and exit. |
| `VERSION` | `false` (Boolean) | Show version of this server. |
| `WEBROOT` | `./www` (String) | Static web server root folder. Only allow change in start parameters. |
| `ZEROTIER` | `false` (Boolean) | Enable ZeroTier functionality for GAN. |
| `ZORAXY_CONFIG_BACKEND` | `local` (String) | Explicit configuration backend: `local` or `postgresql`. PostgreSQL currently covers the migrated HTTP-routing domain. |
| `ZORAXY_NODE_ROLE` | `standalone` (String) | Role reported by health/status endpoints. Use `data-plane` for a PostgreSQL-backed traffic node. |
| `ZORAXY_CONFIGSTORE_MODE` | `disabled` (String) | `disabled`, `control-plane`, or `data-plane`. See `deploy/active-active`. |
| `ZORAXY_CONFIGSTORE_MIGRATION_MODE` | none | Required when configstore is enabled: `verify` or `apply`. |
| `ZORAXY_CONFIGSTORE_DSN_FILE` | none | Path to a Docker secret containing the complete PostgreSQL DSN. Preferred for deployment. |
| `ZORAXY_CONFIGSTORE_DSN` | none | Direct PostgreSQL DSN for disposable test environments only. |
| `ZORAXY_CONFIGSTORE_POLL_INTERVAL` | `1s` | Positive Go duration controlling revision polling. |
| `ZORAXY_CONFIGSTORE_STARTUP_TIMEOUT` | `10s` | Positive Go duration for database/schema preflight and initial activation. |

> [!IMPORTANT]
> Contrary to the Zoraxy README, Docker usage of the port flag should NOT include the colon. Ex: `-e PORT="8000"` for Docker run and `PORT: "8000"` for Docker compose.

### High availability modes

Use [`deploy/ha`](../deploy/ha/README.md) for the production-oriented
active/passive design. Use [`deploy/active-active`](../deploy/active-active/README.md)
only for the staged PostgreSQL migration. A data-plane container performs its
database and schema preflight before opening proxy listeners, ignores legacy
HTTP routing files, and requires an existing valid revision. Never mount the
same writable `/opt/zoraxy/config/` volume into two containers.

For cluster deployments, mount a DSN secret instead of placing credentials in
Compose environment values:

```yaml
services:
  zoraxy:
    environment:
      ZORAXY_NODE_ROLE: data-plane
      ZORAXY_CONFIG_BACKEND: postgresql
      ZORAXY_CONFIGSTORE_MODE: data-plane
      ZORAXY_CONFIGSTORE_MIGRATION_MODE: verify
      ZORAXY_CONFIGSTORE_DSN_FILE: /run/secrets/configstore-dsn
    secrets:
      - configstore-dsn

secrets:
  configstore-dsn:
    file: ./secrets/configstore-dsn
```

### Health endpoints

`GET` and `HEAD` are supported on both unauthenticated endpoints. Restrict port
8000 to the management network even though these responses contain no secrets.

- `/health/live` returns HTTP 200 while the process can serve its management
  handler. It is a process check, not permission to receive proxy traffic.
- `/health/ready` returns HTTP 200 only when every reported check is true;
  otherwise it returns HTTP 503. Checks cover the local database, loaded proxy
  configuration, bound proxy listener and, when enabled, PostgreSQL configstore.

In PostgreSQL data-plane mode verify that `node_role` is `data-plane`,
`config_revision` equals the non-zero `applied_revision`, and
`checks.config_store` is `true`.

### ZeroTier

If you are running with ZeroTier, make sure to add the following flags to ensure ZeroTier functionality:
  
`--cap_add NET_ADMIN` and `--device /dev/net/tun:/dev/net/tun`
`--environment ZEROTIER="true"`

Or for Docker Compose:
```
  cap_add:
    - NET_ADMIN
  devices:
    - /dev/net/tun:/dev/net/tun
  environment:
    ZEROTIER: "true"
```

### Plugins

Zoraxy includes a (experimental) store to download and use official plugins right from inside Zoraxy, no preparation required.
For those looking to use custom plugins, build your plugins and place them inside the volume `/path/to/zoraxy/plugin/:/opt/zoraxy/plugin/` (Adjust to your actual install location).

### Building

To build the Docker image:

- Check out the repository/branch and run the command from the repository root.
- Build with `docker build -f docker/Dockerfile -t zoraxy_build .`.
- To use another image name, replace `zoraxy_build` in the command.
