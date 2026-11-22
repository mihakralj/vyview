# Deployment, Packaging & Operations

> Covers packaging paths onto VyOS routers, the proposed native VyOS CLI config tree, RBAC alignment, hardened systemd units, storage path configuration, and boot sequencing.

---

## 1. Packaging Sequence

| Option | Approach | Trade-Offs & Rollout Stage |
| --- | --- | --- |
| **(a) Custom `.deb` in VyOS image** | Ship `/usr/bin/vyview`, systemd units (`vyview-web`, `vyview-probe`), and default config inside custom LS1046A / x86 VyOS images, **disabled by default**. | **Start here (v0.1–v0.2).** Cleanest host integration, zero container socket friction, survives image builds natively. |
| **(b) VyOS Podman container** | Run vyview inside `set container` with read-only host mounts and separate probe privileges. | Complicates direct netlink, D-Bus, and control-socket access; secondary option only. |
| **(c) Upstream `service dashboard` config node** | Native VyOS XML interface definition + conf-mode script generating `/run/vyview/vyview.toml` and managing systemd units. | **Propose upstream after v0.2 stabilizes** and real-world LS1046A/x86 deployments prove the contract. |

*Boot persistence before Option (c) lands:* Persist the binary and units via custom `.deb` packaging in the image build (`vyos-ls1046a-build`) or via `/config/user-data/` + `/config/scripts/vyos-postconfig-bootup.script`.

---

## 2. Proposed VyOS CLI & RBAC Sketch (Option c)

Illustrative configuration tree (not yet an accepted upstream VyOS interface):

```text
set service dashboard listen-address 192.168.10.1
set service dashboard listen-address fd00:10::1
set service dashboard port 8443
set service dashboard vrf mgmt
set service dashboard certificate vyview-cert
set service dashboard authentication mode pam
set service dashboard authentication allow-group vyattaop
set service dashboard retention profile lean
set service dashboard retention max-age 90d
set service dashboard retention path /var/lib/vyview
set service dashboard retention snapshot-path /mnt/usb/vyview-backup.ring
set service dashboard startup-delay 5s
set service dashboard privacy per-client aggregate
set service dashboard modules ask2
```

### VyOS RBAC Alignment (`operator-group`)
When defining a read-only operator account for vyview (especially for remote/off-router polling):

```text
set system login operator-group vyview-ro command-policy allow 'show'
```

Remote installs pair this operator group with a dedicated **SSH public-key-only** `vyview` user (keys generated and verified on first run; no password auth). See [System Architecture §8](system-architecture.md#8-local-on-router-vs-remote-off-router-deployment-differences) for full Local vs. Remote differences.

---

## 3. Router OS Prerequisites

1. **Conntrack accounting:** `sysctl net.netfilter.nf_conntrack_acct=1` (needed for per-client byte/packet rollups).
2. **Unprivileged ICMP probes:** `sysctl net.ipv4.ping_group_range` covering the `vyview` GID (enables latency probes without `CAP_NET_RAW`).
3. **Service users & supplementary groups:**
   - `vyview` user in `systemd-journal` (journald reads), `vyattacfg-ro` (read-only `/config/archive`), `frrvty` (read-only `vtysh`), plus read access to Kea/VICI/Podman sockets.
   - `vyview-probe` user with `CAP_NET_ADMIN` only.
4. **Persistent storage target:** Mount a dedicated eMMC partition, LV, or external USB drive at `/var/lib/vyview` (or pass `--data-path /mnt/usb/vyview.ring`) so history survives VyOS image upgrades without bloating `/config` backups. If `/config/vyview` is used instead, restrict to the `compact` or `lean` profile and verify backup size impact.
5. **Delayed boot start:** Order units `After=network-online.target vyos-router.service` with optional `--start-delay` (e.g., `5s`) to deconflict with VyOS interface/dataplane initialization.

---

## 4. Hardened systemd Units (Illustrative)

```ini
# /lib/systemd/system/vyview-web.service
[Unit]
Description=vyview Read-Only Dashboard Web & Collector Service
After=network-online.target vyos-router.service vyview-probe.service
Wants=network-online.target

[Service]
User=vyview
Group=vyview
SupplementaryGroups=systemd-journal vyattacfg-ro frrvty
ExecStart=/usr/bin/vyview web --config /run/vyview/vyview.toml --start-delay 5s
CapabilityBoundingSet=
AmbientCapabilities=
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ReadWritePaths=/var/lib/vyview /run/vyview
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM
MemoryMax=96M
CPUQuota=25%
Environment=GOMEMLIMIT=40MiB GOMAXPROCS=2

# /lib/systemd/system/vyview-probe.service
[Unit]
Description=vyview Read-Only Privileged Netlink Probe
After=network-online.target vyos-router.service

[Service]
User=vyview-probe
ExecStart=/usr/bin/vyview probe --socket /run/vyview/probe.sock
AmbientCapabilities=CAP_NET_ADMIN
CapabilityBoundingSet=CAP_NET_ADMIN
NoNewPrivileges=yes
ProtectSystem=strict
RestrictAddressFamilies=AF_UNIX AF_NETLINK
IPAddressDeny=any
```

---

## 5. Build, Release & Upgrade Hygiene

- **Cross-compilation:** `CGO_ENABLED=0 GOOS=linux GOARCH=arm64` (and `amd64`) `go build -trimpath -ldflags='-s -w'`.
- **Reproducible supply chain:** Pin Go and Node versions in `mise.toml`, lock dependencies in `go.sum` and `package-lock.json`, publish SHA-256 checksums (`SHA256SUMS.txt`) and CycloneDX SBOMs (`dist/`).
- **Schema upgrades:** Ring headers carry `format_version`. On a version bump, vyview initializes `vyview.vN.ring` and mounts the previous file read-only until expiry.
- **Quick evaluation on a router (demo binary):** Copy `vyview-linux-arm64` to `/tmp/vyview` and launch a transient unit without modifying persistent router state:
  ```sh
  sudo systemd-run --unit=vyview-demo /tmp/vyview -addr 0.0.0.0:8080 -allow-lan
  ```
