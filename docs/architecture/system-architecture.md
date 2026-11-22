# System Architecture & Router Runtime

> Specifies vyview's runtime topology, Two-Plane Hardware Offload Model, data-gathering pipeline, allowlisted command runner, Datastar/SSE presentation stack, `CGO_ENABLED=0` trade-offs, and Local vs. Remote deployment differences.

---

## 1. Two-File, Two-Process Runtime Topology

vyview is architected first for **on-router execution** on embedded ARM64 (`NXP LS1046A "Mono Gateway"`, Linux 6.18, VyOS 1.5/rolling) and stock `x86_64` VyOS, with zero external runtime dependencies (no Node.js, no Python web server, no external TSDB).

### Files on the Router
1. **`/usr/bin/vyview`** — Single static Go binary (`CGO_ENABLED=0`) containing all collectors, the storage engine, the SSE hub, the HTTP/2 TLS server, the opt-in MCP interface, and all `go:embed`ded frontend assets.
2. **`/var/lib/vyview/vyview.ring`** *(configurable via `--data-path`)* — Single pre-allocated, 4 KiB page-aligned multi-ring binary datastore (see [Storage Format](storage-format.md)).

### Least-Privilege Process Split
The same single binary runs as two hardened systemd units:

| Unit | Command | Privileges | Responsibilities |
| --- | --- | --- | --- |
| `vyview-web.service` | `/usr/bin/vyview web --config /run/vyview/vyview.toml` | **Zero capabilities** (`CapabilityBoundingSet=`, `NoNewPrivileges=yes`, `ProtectSystem=strict`, `SystemCallFilter=@system-service`) | Unprivileged collectors (`/proc`, `/sys`, unprivileged netlink & D-Bus), Two-Plane Offload Reconciler, `vyview.ring` storage engine, state snapshots, SSE hub, HTTPS server (default on LAN), Datastar HTML/signal renderers, recommended Prometheus `/metrics` endpoint (including health checks & alerts), and opt-in MCP server. |
| `vyview-probe.service` | `/usr/bin/vyview probe --socket /run/vyview/probe.sock` | **`CAP_NET_ADMIN` + read-only debugfs** (`IPAddressDeny=any`, `RestrictAddressFamilies=AF_UNIX AF_NETLINK`) | Narrow read-only helper exposing a fixed UNIX-socket RPC surface: `DumpNft`, `DumpConntrack(filter)`, `SubscribeConntrack`, `WGDevices`, `EthtoolModule(iface)`, `DumpXfrm`, `Ask2Netlink` (`ynl --family ask` / `show offload flow`), and `ReadHardwareStats` (QMan/BMan/CAAM/QAT/AF_XDP/`wgfp_rx` + guarded debugfs fallback). |

---

## 2. Two-Plane Hardware Offload Model

General-purpose monitoring tools (`ntopng`, `Netdata`, `OPNsense Insight`, `MikroTik Torch`) sample traffic at CPU-visible taps (`libpcap`, `conntrack`, `netdev`). On LS1046A with DPAA1/ASK2 (or boxes running VPP / AF_XDP / hardware flowtables), accelerated flows bypass the CPU entirely. vyview explicitly models two planes and reconciles them in-process:

```text
Plane A (Acceleration / Silicon)
  ASK2 Generic Netlink (ynl --family ask) + FMan MAC/FIFO (ethtool -S)
  + QMan/BMan pools + VPP stats segment + AF_XDP ZC + CAAM/QAT + wgfp_rx
        │
        ▼
  ┌─────────────────────────────────────────────────────────────────────┐
  │                 In-Process Two-Plane Reconciler                     │
  │  • Per-port & per-flow offload_ratio (flows, packets, AND bytes)    │
  │  • Source-split drops (netdev vs FMan FIFO vs softnet vs BMan)      │
  │  • "Why isn't this flow offloaded?" explanations                    │
  └─────────────────────────────────────────────────────────────────────┘
        ▲
        │
Plane B (Kernel / CPU)
  rtnetlink IFLA_STATS64 (netdev) + CTNETLINK ([OFFLOAD]/[HW_OFFLOAD])
  + nf_flowtable + /proc/net/softnet_stat + /proc/net/xfrm_stat
```

---

## 3. Stage 1 — Data Gathering Pipeline

### Core Scheduling & Resilience Rules
- **Never spawn processes on the fast polling loop:** The `200 ms` (5 Hz) fast loop (interface `IFLA_STATS64`, `/proc/stat`, `/proc/softirqs`, active queue drops) uses strictly in-process netlink calls and `/proc` reads stamped with Go's monotonic clock (`time.Now()`). Slower metrics (`1 s`+) piggyback on every 5th tick (`5 × 200 ms = 1 s`).
- **Concurrency & timeouts:** Single-flight execution per source (never overlapping ticks), at most 4 concurrent external execs globally, strict per-source timeouts (`100 ms` netlink/`/proc`, `2 s` `vtysh`, `5 s` op-mode).
- **Load smoothing & circuit breaking:** Phase-staggered scheduling (`±10%` jitter plus prime/offset start delays) prevents `5 s`, `10 s`, `30 s`, and `60 s` collectors from colliding on the same second. Repeated failures back off exponentially up to `5 min` with circuit-breaker retry limits.
- **Snapshot isolation:** Parsers are JSON-first with release-specific text fallbacks tested against golden captures. Normalized, redacted state is published via copy-on-write snapshots feeding the ring store, event detector, search index, and SSE hub.

### 3.1 Go Collector Libraries (Compiled Into `vyview`)

| Subsystem | Go Package (Preferred) | Kernel / OS Mechanism | Privilege | Cadence |
| --- | --- | --- | --- | --- |
| Links, addresses, routes, neighbors | `vishvananda/netlink` / `mdlayher/netlink` | `RTM_GETLINK` (`IFLA_STATS64`), `RTM_GETADDR/ROUTE/NEIGH`, plus event-driven `rtnetlink` multicast (`Link/Addr/NeighSubscribe`) | `vyview-web` | `200 ms` stats (monotonic $\Delta t$), event-driven flaps, `5–30 s` routes |
| nftables rules, sets, flowtables | `google/nftables` | Netlink `NFT_MSG_GET*` (read-only) | `vyview-probe` | `5 s` counters, `10–60 s` flowtables/sets |
| Conntrack, NAT & per-client accounting | `ti-mo/conntrack` | `CTNETLINK` (`IPCTNL_MSG_CT_GET*` + event-driven `NEW`/`DESTROY` stream; full dump only while open) | `vyview-probe` | Event stream + `5 s` stats, `10 s` Top-K rollup, `5 s` dump on demand |
| QoS qdiscs, classes, CAKE xstats | `florianl/go-tc` | Netlink `RTM_GETQDISC`, `RTM_GETTCLASS` | `vyview-web` | `5 s` |
| WireGuard peers & fast-path RX | `wireguard/wgctrl` + ASK2 `wgfp_rx` | Generic netlink `WG_CMD_GET_DEVICE` (private keys dropped at boundary); reconciled with `wgfp_rx` | `vyview-probe` | `5 s` |
| IPsec SAs & crypto offload | `vishvananda/netlink` (`XfrmStateList`) + `/proc/net/xfrm_stat` + CAAM | `NETLINK_XFRM` + `/proc/net/xfrm_stat` + `/proc/crypto` & CAAM job counters | `vyview-probe` | `10 s` |
| NIC link modes, FMan MAC stats, SFP+ DOM | `mdlayher/ethtool` (+ ioctl fallback) | `ETHTOOL_GSTATS` (incl. DPAA/FMan MAC & RX FIFO overflows), `GLINKMODES`, SFF-8472/8636 EEPROM | `vyview-probe` (EEPROM) | `1–5 s` MAC/FIFO, `300 s` DOM (slow I2C) |
| systemd units & VRRP HA state | `godbus/dbus/v5` | System D-Bus read-only `ListUnits` + event-driven `keepalived` VRRP signals (zero polling) | `vyview-web` | Event-driven (VRRP), `30 s` (units) |
| Built-in latency & DNS probes | Go `net` + `x/net/icmp` | Unprivileged ICMP datagram sockets (`ping_group_range`) & UDP/DNS, hard rate-limited (`≤ 1 probe/target/10 s`, max `8` targets) | `vyview-web` | `10 s` RTT, `60 s` DNS |
| Acceleration (ASK2, VPP, AF_XDP, CAAM/QAT) | `mdlayher/genetlink` (`ynl --family ask`) + VPP stats + `XDP_STATISTICS` | **Prefer ASK2 Generic Netlink (`ynl --family ask` / `show offload flow`) over unstable `fe_flow` debugfs** (debugfs kept only as guarded legacy fallback); QMan/BMan, `wgfp_rx`, CAAM/QAT, AF_XDP ZC, L2/TSN protocol detection | `vyview-probe` | `5 s` (hidden when absent) |
| OT time & ring telemetry (PTP / MRP) | `bridge mrp` netlink + `pmc` / `phc2sys` | Event-driven MRP ring transitions (`< 2 s` UI target, `100%` capture); PTP clock offset & sync state | `vyview-web` | Event-driven (MRP), `1 s` (PTP) |
| Direct `/proc`, `/sys`, and `sock_diag` | Go stdlib (pooled `[]byte` buffers) + `NETLINK_SOCK_DIAG` | `/proc/{stat,loadavg,meminfo,softirqs,interrupts,diskstats,cmdline,modules,pressure/*}`, `/proc/net/{softnet_stat,xfrm_stat}`, `/proc/<pid>/stat\|status` (no `ps`), `NETLINK_SOCK_DIAG` (no `ss`), `/run/utmp` / `sshd` journal (no `w`), `/sys/class/hwmon/*` (TMP431), **`/sys/block/mmcblk*/device/{life_time,pre_eol_info}`** (no `mmc-utils`) | `vyview-web` | `200 ms` CPU/softirq, `1 s`–`1 h` thermals/mem/eMMC |

### 3.2 Local Control Sockets

Every local socket's reachability, last success timestamp, and error count are surfaced directly in **Services** and **Health & Alerts** so localized daemon failures appear explicitly rather than as silent blanks:

| Daemon | Socket / Protocol | Read-Only Operations Allowed | Cadence |
| --- | --- | --- | --- |
| **Kea DHCPv4 / DHCPv6** | `/run/kea/dhcp4-ctrl.sock`, `dhcp6-ctrl.sock` (JSON) | `lease4-get-all`, `lease6-get-all`, `stat-lease4-get`, `statistic-get-all` | `30 s` |
| **strongSwan IPsec** | `/run/charon.vici` (pure-Go VICI client) | `list-sas` only (no initiate/terminate) | `10 s` |
| **Podman containers** (`set container`) | `/run/podman/podman.sock` or `/run/user/<uid>/podman/podman.sock` (HTTP-over-UNIX, `GET` only) | `GET /libpod/containers/json`, `GET /libpod/containers/{id}/stats?stream=false` (read-only status, image, uptime, CPU/mem/net; no start/stop/exec) | `10–30 s` |
| **Suricata** *(optional)* | `eve.json` stats tail or UNIX socket | `dump-counters` (verdict summaries only) | `10 s` |

### 3.3 Allowlisted Read-Only Executables (No Shell)

Where no kernel netlink or control socket exists, vyview runs only compile-time allowlisted absolute `argv` templates via `exec.CommandContext` — never `/bin/sh`. Placeholders are strongly typed and validated against the current state model (e.g., interface names match `^[a-z][a-z0-9.\-]{0,14}$`, archive revisions are bounded integers), with `LANG=C`, fixed `PATH`, output caps (`4–16 MiB`), and strict timeouts.

| Binary (Present in VyOS) | Invocation Pattern | Format | Cadence |
| --- | --- | --- | --- |
| **FRR `vtysh`** | `/usr/bin/vtysh -c 'show bgp vrf all summary json' ...` (batched into one exec) | JSON | `5 s` BFD, `30 s` BGP/OSPF/RIB |
| **systemd `journalctl`** | `/usr/bin/journalctl -o json -f` (one long-lived child) + `--list-boots -o json` | NDJSON | Event-driven + `1 h` boots |
| **PowerDNS `rec_control`** | `rec_control get-all` | Key-value | `30 s` |
| **Chrony `chronyc`** | `chronyc -c sources`, `-c tracking`, `-c clients` | CSV | `60 s` |
| **LLDP `lldpcli`** | `lldpcli -f json show neighbors details` | JSON | `60 s` |
| **VyOS op-mode & config** | `show configuration json`, `/config/archive/config.boot.N.gz` (`inotify`), `vyos.opmode` | JSON / configtree | On commit / `300 s` / on demand |
| **`mtr`** *(scheduled)* | `mtr --json` (fixed configured targets only) | JSON | `15 min` |

Illustrative allowlist entry format (compiled into binary; paths verified per VyOS release):

```yaml
- id: frr.bgp.summary
  argv: ["/usr/bin/vtysh", "-c", "show bgp vrf all summary json"]
  timeout: 2s
  max_output: 4MiB
  parser: frr.bgp_summary.v1
- id: ethtool.module
  argv: ["/usr/sbin/ethtool", "-m", "{iface:known_ethernet}"]
  timeout: 3s
  parser: ethtool.sff8472.text.v1
```

> **Zero extra OS packages required for hardware health or process lists:** eMMC `EXT_CSD` wear (`life_time`, `pre_eol_info`) comes straight from Linux sysfs (`mmc-utils` is **not** needed). Processes, sockets, and sessions come from `/proc`, `NETLINK_SOCK_DIAG`, and `journald` (no `ps`, `ss`, or `w`).
>
> **No VyOS HTTP API key in on-box mode:** VyOS API keys grant full read/write/reboot access and cannot be scoped read-only.

---

## 4. Stage 2 — Storage & Metrics Export

All persistent telemetry, series metadata, curated events, alert transitions, diagnostic probes, and an optional `T1` clean-shutdown snapshot live in **`vyview.ring`** — a single pre-allocated, 4 KiB page-aligned multi-ring file (`~21–32 MB/day` host writes in full profile, `~14.5 MB/day` lean, `< 5 MB/day` compact). Full binary geometry, tier math (`200 ms → 5 s → 60 s → 5 min → 1 h`), cross-tier statistical aggregation, and bounded `≤ 4 MiB` RAM tier warming are specified in [Storage Format (`vyview.ring`)](storage-format.md).

**Recommended Prometheus `/metrics` endpoint (KISS default):** vyview exposes an authenticated read-only `/metrics` endpoint exporting **all collected router telemetry plus vyview's own health checks and alerts**. Operators can point Prometheus + Alertmanager (optionally running as a VyOS Podman container) at this single endpoint without deploying separate `node-exporter`, `frr-exporter`, or `blackbox-exporter` daemons. Optional remote-write is supported when an off-router TSDB exists.

---

## 5. Stage 3 — Presentation & Streaming (Datastar + Browser)

All frontend assets are built at CI time, content-hashed (`/a/<hash>.<ext>`), pre-compressed (Brotli + Gzip), and embedded via `go:embed`. The router serves everything offline on the LAN with zero CDN calls.

| Layer | Technology | Runtime Role | Est. Brotli Payload |
| --- | --- | --- | --- |
| **Hypermedia** | **Datastar v1.0.x** (self-hosted ES module) + **Datastar Go SDK** | Single SSE stream (`GET /sse/{page}`) per open page; DOM morphing by stable element ID (`datastar-patch-elements`), log append/remove, SVG namespace patching, and private `_chart.*` signals (`datastar-patch-signals`). | `11.0 KB` |
| **Templates** | **`templ`** (preferred, compile-time typed) or **`html/template`** *(decided in MVP Week 1)* | Complete server-rendered first paint (`< 300 ms` LAN, works without JS); fragment hash diffing in the SSE hub; contextual auto-escaping; static lint banning dynamic `data-*` attributes. | `9.5 KB` (Overview HTML) |
| **Primary charts** | **uPlot** inside `<vy-chart>` Web Component | Decouples `60 fps` `requestAnimationFrame` X-axis scrolling (`250–300 ms` render lag + leading-edge interpolation) from the `5 Hz` telemetry stream. Synced crosshairs, brush-to-zoom, anomaly bands, event annotations, compare-to-previous overlay. Server caps queries to `1,000–1,500` pts via M4/LTTB. | `19.5 KB` |
| **Inline graphics** | **Server-rendered SVG** | Zero-JS `<polyline>` sparklines (~300 B each), small heatmaps (`≤ 2,000` cells), BMan gauges, and Sugiyama-layered L2/L3 topology map (live utilization patched via SVG `stroke`). | `0 KB` client JS |
| **Heavy visuals** *(lazy)* | **Apache ECharts** (custom modular build: Sankey + Heatmap + Graph) | Lazy-loaded only on Traffic & Clients (Sankey), complex Topology graphs, and large heatmaps. *(Cytoscape.js/D3 were evaluated and declined to avoid shipping a second graph library.)* | `118.0 KB` *(lazy)* |
| **Client helpers** | Vanilla **`app.js`** | `<vy-chart>`, `<vy-palette>` (`Ctrl/Cmd-K`), `<vy-range>`, keyboard shortcut sequencer, `1 s` local stale-age ticker. | `7.0 KB` |
| **CSS & tokens** | **UnoCSS** (`presetWind3`, CI-only CLI) + `tokens.css` | Dark-first color-blind-safe palette, light theme, and 3 densities switched purely via CSS variables. | `6.0 KB` |
| **Icons** | Subsetted SVG `<symbol>` sprite (`~64` symbols) | Private builds (`-tags fapro`): **Font Awesome Pro 7.3.1 Sharp Light**. Public/upstream builds (default): **Font Awesome Free 7.x** with identical symbol IDs. Zero runtime FA JS/webfonts. | `13.8 KB` Pro / `11.8 KB` Free |
| **Typography** | Self-hosted WOFF2 Latin subsets (**IBM Plex Sans** + **IBM Plex Mono**) | Local WOFF2 fonts with `tabular-nums`. Optional `-tags systemfonts` drops initial payload from `~139.8 KB` to `~69.8 KB` (budget `< 150 KB`). | `46.0 KB` + `24.0 KB` *(opt.)* |
| **PWA shell** | Manifest + minimal `sw.js` | Caches immutable hashed assets only (never live data/SSE); network-first HTML with offline fallback. | `3.0 KB` + `1.6 KB` *(lazy)* |

### SSE Hub & Transport Guarantees
- **HTTP/2 over TLS** avoids HTTP/1.1's 6-connection-per-host SSE ceiling (`≤ 8` streams/session, `≤ 32` global).
- **No compression on `text/event-stream`** (prevents proxy/buffer flush stalls; live Overview deltas are ~5–10 KB/s). Static assets serve pre-compressed `.br`/`.gz`.
- **Shared subscriptions & backpressure:** Subscriptions key on `(session, page, entities, tier, filter_hash)`, coalescing identical renders. Per-client send queues are bounded (e.g., `64` events / `256 KiB`); overflow drops intermediate deltas and schedules a clean snapshot, and stalled clients disconnect after `30 s`.
- **Keepalive & reconnect:** `15 s` SSE keepalive; exponential client retry (`1 s`, factor `2`, max `30 s`) with a full snapshot on every reconnect. Hidden-tab suspension is configurable.

---

## 6. Static Binary (`CGO_ENABLED=0`) Trade-Offs

| Feature | C / `cgo` Library | Impact on Static Cross-Compile | Resolution for vyview |
| --- | --- | --- | --- |
| **journald streaming** | `sdjournal` (`libsystemd.so`) | Breaks `CGO_ENABLED=0` and couples to host `glibc`/`libsystemd` ABI across VyOS upgrades. | Use one long-lived allowlisted `journalctl -o json -f` child (~0.05 ms/event). Keeps binary 100% static. |
| **VyOS PAM login** | `libpam.so` | Direct `pam_authenticate` requires `cgo` and host `libpam`. | Validate in MVP Week 1: prefer pure-Go auth (`argon2id` + VyOS PKI mTLS + forward-auth) plus an optional tiny OS PAM verifier subprocess to keep `CGO_ENABLED=0`, or adopt an `arm64` cross-CGO build only if in-process PAM is mandatory. |

---

## 7. On-Router Runtime vs. CI-Only Build Toolchain

| Runs on the Router | CI / Developer Machine Only (Never on Router) |
| --- | --- |
| `/usr/bin/vyview` (`vyview-web` + `vyview-probe`) | Go compiler, `templ` CLI, `govulncheck`, fuzzers |
| `/var/lib/vyview/vyview.ring` (`--data-path`) | Node.js, npm, UnoCSS CLI, ECharts modular bundler |
| `/run/vyview/` (`tmpfs`: `5 s` hot ring, `probe.sock`, rendered `vyview.toml`) | Font Awesome icon subsetter & Pro leak-guard CI check |
| Existing VyOS/Linux interfaces (netlink, `/proc`, `/sys`, FRR, Kea, VICI, Podman, `journalctl`) | Playwright + `axe-core`, `.deb` packager, CycloneDX SBOM |

*Explicitly never on the router:* Node.js, SQLite, PostgreSQL, InfluxDB, ClickHouse, Grafana, custom Python web servers, `/bin/sh` execution, CDN links, or external font services.

---

## 8. Local (On-Router) vs. Remote (Off-Router) Deployment Differences

On-router execution is the **primary architecture**. Off-router (remote) execution on a management host is an optional secondary mode with clear trade-offs:

| Dimension | Local (On-Router — Primary) | Remote (Off-Router — Optional) |
| --- | --- | --- |
| **Install & Config** | Single binary + `.deb` or `/config/user-data/` + systemd units; config generated from VyOS CLI (`service dashboard`) or `/config/vyview/vyview.toml`. | Explicit `vyview.toml` **required** (target host, TLS CA/cert paths, credentials/SSH key path, poll cadences). |
| **Data Collection** | Direct `200 ms` kernel netlink (`IFLA_STATS64`, `CTNETLINK`, `NFT`, `wgctrl`, ASK2 netlink), `/proc`, `/sys`, and local UNIX sockets. Zero VyOS HTTP API key needed. | Restricted VyOS HTTP API (`/retrieve`, allowlisted `/show`, `GET /info`) and/or a dedicated pubkey-only SSH `vyview` user bound to `operator-group vyview-ro` (`allow 'show'`). No `200 ms` kernel netlink or raw ASK2/BMan stats unless an on-router agent exists. |
| **TLS Certificates** | VyOS PKI (`pki certificate <name>`) with automatic reload; LAN HTTPS by default. | Operator-managed TLS (internal CA, Let's Encrypt/ACME, or fronting reverse proxy such as Caddy/nginx). |
| **Security & Auth** | Unprivileged `vyview-web` + `CAP_NET_ADMIN` `vyview-probe`; local PAM / VyOS operator group / mTLS / `argon2id`. | Dedicated pubkey-only SSH key (generated & tested on first run) + compile-time outbound HTTP API allowlist blocking all `/configure` and mutation paths; root-owned helper-only API key if HTTP API is used. |
| **Storage & Wear** | Single-file `vyview.ring` (eMMC or `--data-path` USB) + RAM `T0`/`T1` tuned for `< 50 MB/day`. | Standard host disk/SSD; higher retention and passive snapshots have negligible wear impact. |
