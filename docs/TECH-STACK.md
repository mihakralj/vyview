# vyview Technology Stack & Router Runtime Requirements

> Derived from *VyOS View-Only Dashboard Design Brief v1.1 (FA Pro, 17 Nov 2026)* and refined for a **two-file router footprint**: **1 static Go binary** (`/usr/bin/vyview`) + **1 single-file multi-ring datastore** (`/var/lib/vyview/vyview.ring`).

This document specifies every technology required **on the VyOS router** across the three pipeline stages (**Gather → Store → Present via Datastar**), the underlying Linux/VyOS OS prerequisites, critical static-binary (`CGO_ENABLED=0`) engineering trade-offs, and the build-time/CI tools that **never** touch the router.

---

## 1. Runtime Architecture on the Router

`vyview` ships as **one static Go binary** (`linux/arm64` for NXP LS1046A "Mono Gateway" or `linux/amd64` for stock x86 VyOS) containing all collectors, the storage engine, the SSE hub, the HTTP/2 TLS server, and all `go:embed`ded frontend assets.

For least privilege (Brief §11.3, Option B), the **same single binary** is executed as two hardened systemd services:

| Unit | Command | Capabilities / Privileges | Role |
| --- | --- | --- | --- |
| `vyview-web.service` | `/usr/bin/vyview web --config /run/vyview/vyview.toml` | **Zero capabilities** (`CapabilityBoundingSet=`, `NoNewPrivileges=yes`, `ProtectSystem=strict`, `SystemCallFilter=@system-service`) | Unprivileged collectors (`/proc`, `/sys`, unprivileged netlink/D-Bus multicast), Two-Plane Offload Reconciler, single-file multi-ring datastore, state model, SSE hub, HTTPS server, Datastar renderers |
| `vyview-probe.service` | `/usr/bin/vyview probe --socket /run/vyview/probe.sock` | **`CAP_NET_ADMIN` + read-only debugfs** (`IPAddressDeny=any`, `RestrictAddressFamilies=AF_UNIX AF_NETLINK`) | Minimal read-only privileged helper exposing a fixed UNIX-socket RPC surface: `DumpNft`, `DumpConntrack(filter)`, `SubscribeConntrack`, `WGDevices`, `EthtoolModule(iface)`, `DumpXfrm`, `ReadAsk2Debugfs` (`fe_flow`, `wgfp_rx`, `qbman`, `caam`) |

---

## 2. Stage 1 — Data Gathering Technologies (On Router)

The core collection rule (Brief §9.3, §12.1) is **never spawn processes on the fast polling loop**. Fast telemetry (**200 ms / 5 Hz** for interface RX/TX bytes & packets, CPU/softirq, and queue drops; slow metrics piggybacked on every 5th tick at **1 s**+) comes exclusively from direct kernel netlink calls, event-driven netlink/D-Bus multicast subscriptions, and `/proc` + `/sys` reads stamped with monotonic $\Delta t$; external CLI tools run only from a compile-time allowlist at 30 s–300 s or on demand.

**Two-Plane Offload Accounting Model (LS1046A / ASK2):** Because hardware-offloaded flows bypass CPU taps (`netdev`, `conntrack`, `libpcap`), `vyview` explicitly separates **Plane A (Hardware: FMan `ethtool -S` MAC counters, ASK2 `fe_flow` debugfs, QMan/BMan pools, CAAM crypto jobs, `wgfp_rx`)** from **Plane B (Kernel: `IFLA_STATS64` netdev, `CTNETLINK` + `[OFFLOAD]`/`[HW_OFFLOAD]` bits, `softnet_stat`, `xfrm_stat`)**, and reconciles them in-process to expose `offload_ratio` per port and per flow plus source-split drop counters.

### 2.1 Go Collector Libraries (Compiled Into `vyview`)

| Subsystem | Go Package (Preferred) | Kernel / OS Mechanism | Privilege Scope | Cadence |
| --- | --- | --- | --- | --- |
| Interfaces, links, addresses, routes, ARP/NDP | `github.com/vishvananda/netlink` and/or `github.com/mdlayher/netlink` | `RTM_GETLINK` (`IFLA_STATS64`), `RTM_GETADDR`, `RTM_GETROUTE`, `RTM_GETNEIGH`, plus **event-driven `rtnetlink` multicast** (`LinkSubscribe`/`AddrSubscribe`/`NeighSubscribe`) | Unprivileged (`vyview-web`) | `200 ms` (5 Hz live stats, monotonic Δt), event-driven flaps, 5–30 s (routes) |
| nftables ruleset, sets, flowtables & counters | `github.com/google/nftables` | Netlink `NFT_MSG_GET*` (read-only list calls) | `CAP_NET_ADMIN` (`vyview-probe`) | 5 s (counters), 10–60 s (flowtables/sets) |
| Conntrack stats, NAT translations, per-client accounting | `github.com/ti-mo/conntrack` | Netlink `CTNETLINK` (`IPCTNL_MSG_CT_GET*` + **event-driven `NEW`/`DESTROY` stream** for churn/accounting; full table dump on demand) | `CAP_NET_ADMIN` (`vyview-probe`) | Event stream + 5 s stats, 10 s Top-K sketch rollup, 5 s dump while open |
| QoS qdiscs, classes, CAKE xstats | `github.com/florianl/go-tc` | Netlink `RTM_GETQDISC`, `RTM_GETTCLASS` | Unprivileged (`vyview-web`) | 5 s |
| WireGuard interfaces, peers & fast-path RX | `golang.zx2c4.com/wireguard/wgctrl` + ASK2 `wgfp_rx` | Generic netlink (`WG_CMD_GET_DEVICE`); private keys dropped at collector boundary; reconciled with `wgfp_rx` when ASK2 present | `CAP_NET_ADMIN` (`vyview-probe`) | 5 s |
| IPsec SAs & crypto offload (`xfrm` + CAAM) | `vishvananda/netlink` (`XfrmStateList`) + `/proc/net/xfrm_stat` + CAAM debugfs | `NETLINK_XFRM` dump + `/proc/net/xfrm_stat` + `/proc/crypto` & CAAM job counters (verifies HW crypto vs SW fallback) | `CAP_NET_ADMIN` (`vyview-probe`) | 10 s |
| NIC link modes, DPAA/FMan MAC stats, SFP+ DOM EEPROM | `github.com/mdlayher/ethtool` (+ ioctl fallback) | Ethtool netlink (`ETHTOOL_GSTATS` incl. FMan MAC/RX FIFO overflows, `GLINKMODES`) / SFF-8472 & SFF-8636 decode | `CAP_NET_ADMIN` for EEPROM (`vyview-probe`) | 1–5 s MAC/FIFO stats, 300 s DOM (slow I2C) |
| systemd unit status & VRRP HA state | `github.com/godbus/dbus/v5` | System D-Bus read-only `ListUnits` + **event-driven `keepalived` VRRP D-Bus signals** (master/backup transitions, zero polling) | Unprivileged (`vyview-web`) | Event-driven (VRRP), 30 s (units) |
| Built-in latency/DNS probes | Go `net` + `golang.org/x/net/icmp` | Unprivileged ICMP datagram sockets (`ping_group_range`) & UDP/DNS | Unprivileged (`vyview-web`) | 10 s (RTT), 60 s (DNS) |
| ASK2 / DPAA1 hardware plane *(LS1046A)* | Pure-Go debugfs/sysfs parser (`vyview-probe`) | `fe_flow` debugfs (capped entries, joined to `conntrack` tuples), `fsl_dpa`/`qbman` BMan/QMan pool & FQ congestion, `wgfp_rx`, CAAM stats (capability-detected; hidden when absent) | Read-only debugfs (`vyview-probe`) | 5 s (FMan/QMan/BMan/fe_flow) |
| Direct kernel virtual files | Go stdlib (`os`, pre-allocated `[]byte` buffers) | `/proc/stat`, `/proc/loadavg`, `/proc/meminfo`, `/proc/softirqs`, `/proc/interrupts`, `/proc/net/softnet_stat`, `/proc/net/xfrm_stat`, `/proc/pressure/*`, `/proc/diskstats`, `/proc/cmdline`, `/proc/modules`, `/proc/sys/kernel/tainted`, `/proc/<pid>/stat\|status`, `/sys/class/net/*/statistics/*`, `/sys/class/hwmon/*` (TMP431 CPU/PCB/SFP+), `/sys/class/thermal/*`, `/sys/block/mmcblk*/device/life_time\|pre_eol_info`, `/sys/firmware/devicetree/base/model` | Unprivileged (`vyview-web`) | `200 ms` (CPU/softirq), `1 s`–`1 h` (softnet/thermals/mem/flash) |

### 2.2 Local Control Sockets (Already Running on VyOS)

| Daemon | Socket / Protocol | Read-Only Commands Allowed | Cadence |
| --- | --- | --- | --- |
| **Kea DHCPv4 / DHCPv6** | UNIX control socket (`/run/kea/dhcp4-ctrl.sock`, `dhcp6-ctrl.sock`), JSON | `lease4-get-all`, `lease6-get-all`, `stat-lease4-get`, `statistic-get-all` (client rejects any non-allowlisted command) | 30 s |
| **strongSwan IPsec** | VICI UNIX socket (`/run/charon.vici`), binary VICI protocol (pure Go client) | `list-sas` only (no initiate/terminate) | 10 s |
| **Suricata** *(optional)* | `eve.json` stats tail or UNIX socket | `dump-counters` (verdict summaries only) | 10 s |

### 2.3 Allowlisted Read-Only Router Executables (No Shell)

Where no netlink/socket interface exists, `vyview` invokes a compile-time allowlist via `exec.CommandContext` with an explicit `argv` slice, minimal environment (`LANG=C`, fixed `PATH`), capped stdout (`4–16 MiB`), strict timeout, and placeholders validated against the current state model (Brief §9.3):

| Router Binary (Already in VyOS) | Invocation Pattern | Output Format | Cadence |
| --- | --- | --- | --- |
| **FRR `vtysh`** | `/usr/bin/vtysh -c 'show bgp vrf all summary json' -c 'show ip route summary json' ...` (batched into one exec per tick) | Structured JSON | 5 s (BFD), 30 s (BGP/OSPF/RIB) |
| **systemd `journalctl`** | `/usr/bin/journalctl -o json -f` (single long-lived child process; keeps `CGO_ENABLED=0`) + `--list-boots -o json` | NDJSON stream | Event-driven + 1 h boots |
| **PowerDNS Recursor `rec_control`** | `rec_control get-all` | Key-value text | 30 s |
| **Chrony `chronyc`** | `chronyc -c sources`, `-c tracking`, `-c clients` | CSV | 60 s |
| **LLDP `lldpcli`** | `lldpcli -f json show neighbors details` | JSON | 60 s |
| **VyOS op-mode & config** | `show configuration json` / `/config/archive/config.boot.N.gz` (read-only + inotify) / `vyos.opmode` raw/JSON | JSON / configtree | On commit / 300 s / on demand |
| **`mtr`** *(scheduled diagnostics)* | `mtr --json` (fixed config targets only; never user-entered) | JSON | 15 min |
| **ASK2 `ask2ctl`** *(optional LS1046A module)* | `ask2ctl flows --json`, `qos --json`, `wg --json`, `acl --json` + DPAA1 sysfs/debugfs (TBD by ASK2) | JSON / sysfs | 1 s–5 s (hidden when absent) |

> **VyOS HTTP API (`service https api`) is NOT used in on-box mode.** VyOS API keys grant full read/write/reboot access and cannot be scoped read-only (Brief §9.4).

### 2.4 Router OS & Kernel Prerequisites

To support all collectors cleanly on VyOS 1.4+ / 1.5 rolling (Linux 6.18 on LS1046A):

1. **Conntrack accounting enabled:** `sysctl net.netfilter.nf_conntrack_acct=1` (required for per-client byte/packet rollups in Traffic & Clients).
2. **Unprivileged ICMP echo sockets:** `sysctl net.ipv4.ping_group_range` covering the `vyview` GID (enables built-in latency probes without `CAP_NET_RAW`).
3. **System users & read-only group memberships:**
   - `vyview` user in `systemd-journal` (journald reads), `vyattacfg-ro` (new read-only group for `/config/archive`), `frrvty` (read-only `vtysh`), plus read access to Kea/VICI sockets.
   - `vyview-probe` user with `AmbientCapabilities=CAP_NET_ADMIN` only.
4. **TLS certificates:** Issued via VyOS PKI (`pki certificate <name>`), TLS 1.3 preferred (TLS 1.2 AEAD allowed), automatic hot-reload on cert rotation.
5. **VRF & LAN binding:** Socket bound to explicit LAN/mgmt VRF address via `SO_BINDTODEVICE` or `ip vrf exec mgmt` — startup refuses `0.0.0.0`/`::` unless protected by input firewall rules.
6. **Optional storage mount:** Dedicated eMMC partition or LV mounted at `/var/lib/vyview` so `/var/lib/vyview/vyview.ring` survives VyOS image upgrades without bloating `/config` backups (Brief §10.5).

---

## 3. Stage 2 — Data Storage Technology (Single-File Multi-Ring Datastore)

Instead of an external TSDB, SQLite WAL (1,109 MB/day host writes, 24.7% eMMC wear), bbolt (CoW amplification), or multiple separate RRD/event files, `vyview` uses a **custom, pure-Go, single-file multi-ring binary datastore** (full specification in [STORAGE-FORMAT.md](STORAGE-FORMAT.md)):

| Tier | Resolution | Span | Location | Size (Full Profile) | Flash Wear |
| --- | --- | --- | --- | --- | --- |
| **T0 Live Ring** | `200 ms` (5 Hz) | `6 min 50 s` (`2,048` pts) | Process RAM (contiguous 4 KiB-aligned `float32` slab) | `~1.5 MiB` | `0` |
| **T1 Hot Ring** | `5 s` (`25 × T0`) | `24 h` (`17,280` pts) | RAM / `tmpfs` (`/run/vyview`), optional zstd snapshot on clean shutdown | `~50 MiB` *(or `~6 MiB` compact)* | `0` during operation |
| **T2 Warm Ring** | `60 s` (`12 × T1`) | `14 d` (`20,160` slots) | **`vyview.ring`** (region inside single file) | `~236 MiB` (page-aligned) | 1 contiguous write / 60 s |
| **T3 Cool Ring** | `300 s` (`5 × T2`) | `90 d` (`25,920` slots) | **`vyview.ring`** (region inside single file) | `~304 MiB` (page-aligned) | 1 contiguous write / 5 min |
| **T4 Cold Ring** | `3600 s` (`12 × T3`) | `2 y` (`17,520` slots) | **`vyview.ring`** (region inside single file) | `~205 MiB` (page-aligned) | 1 contiguous write / 1 h |
| **Event / Alert / Probe Ring** | Event-driven | `30 d` events / `90 d` alerts & probes | **`vyview.ring`** (circular region inside single file) | `16 MiB` | Batched ≤ 1 / 5 s |
| **T1 Shutdown Snapshot** *(opt.)* | Snapshot | Latest `24 h` hot tier | **`vyview.ring`** (reserved zstd slot inside single file) | `16 MiB` | Written only on `SIGTERM` |

**Why single-file + time-major + `fallocate` + `fdatasync` wins on eMMC:**
- **Zero file growth, zero fragmentation, zero compaction:** Pre-allocated once; offsets are deterministic arithmetic.
- **Minimal write amplification:** One contiguous 12 KiB (3 × 4 KiB pages) `pwrite` per 60 s flush + single `fdatasync()` (no file-size/inode metadata churn, no per-minute superblock write because each slot is self-describing with `(timestamp, slot_seq, valid_bitset, crc32c)`).
- **Host writes:** `~21–32 MB/day` (`< 50 MB/day` budget; `< 0.71%` of a 16 GB MLC eMMC's 10-year endurance at 3× WAF).
- **Simple operations:** Backing up, wiping, or inspecting all persistent `vyview` state means handling **one file**.
- **Optional external integration (Brief §10.6):** Authenticated read-only `/metrics` (Prometheus scrape) and optional remote-write to Prometheus/VictoriaMetrics when an off-router server exists.

---

## 4. Stage 3 — Presentation & Streaming Technologies (Datastar + Browser)

All frontend assets are pre-built at CI time, pre-compressed (Brotli + Gzip), content-hashed (`/a/<hash>.<ext>`), and compiled into the Go binary via `go:embed`. The router serves **everything offline on the LAN** with zero CDN requests.

| Layer | Technology | Role & Runtime Characteristics | Compressed Payload (Brotli) |
| --- | --- | --- | --- |
| **Hypermedia framework** | **Datastar v1.0.x** (self-hosted ES module) + **Datastar Go SDK** (`github.com/starfederation/datastar-go/datastar`) | Single SSE stream (`GET /sse/{page}`) per open page; DOM morphing by stable element ID (`datastar-patch-elements`), append/remove for log tails, SVG namespace patching for topology/sparklines, and private signals (`datastar-patch-signals` → `_chart.*`) for charts. | `11.0 KB` |
| **Server HTML templates** | Go **`templ`** (compiled Go functions, preferred) or stdlib **`html/template`** | Full server-rendered first paint (`< 300 ms` LAN, works without JS); fragment-level hash diffing in the SSE hub so only changed fragments ship; contextual auto-escaping; static lint forbids dynamic `data-*` attributes. | `9.5 KB` (Overview HTML + inline SVG sparklines) |
| **Primary charts** | **uPlot** (`uPlot.min.js` + `uPlot.min.css`) wrapped in `<vy-chart>` plain Web Component | Canvas 2D time-series charts decoupled from the 5 Hz telemetry tick: `requestAnimationFrame` drives smooth **60 fps** X-axis scrolling with a `250–300 ms` (`1.25×–1.5×`) jitter-buffer render lag and leading-edge interpolation (plus optional light causal EMA smoothing on live rates and peak `max` envelope on rolled-up tiers). Synced crosshair (`cursor.sync.key = page`), brush-to-zoom, anomaly bands, event annotation plugin, compare-to-previous overlay. Server caps queries to ~1,000–1,500 points via M4/LTTB downsampling. | `19.5 KB` |
| **Inline graphics** | **Server-rendered SVG** | Zero-JS inline `<polyline>` sparklines (~300 B each), small heatmaps (`≤ 2,000` cells), BMan gauges, and Sugiyama-layered L2/L3 topology map (live link utilization patched via SVG `stroke` updates). | `0 KB` client JS |
| **Complex visualizations** *(lazy)* | **Apache ECharts** (custom modular build: Sankey + Heatmap + Graph only) | Lazy-loaded **only** on Traffic & Clients (Sankey), Topology, and System/Dataplane large heatmaps. | `118.0 KB` *(lazy)* |
| **Custom Web Components** | Vanilla **`app.js`** (plain Web Components; Datastar Rocket not required) | `<vy-chart>`, `<vy-palette>` (Ctrl/Cmd-K command palette), `<vy-range>` (time-range picker), keyboard shortcut sequencer, 1 s local stale-age ticker. | `7.0 KB` |
| **CSS & Design Tokens** | **UnoCSS** (`presetWind3`, build-time CLI only) + CSS custom properties (`tokens.css`) | Dark-first color-blind-safe palette + light theme + 3 densities (`comfortable`, `compact`, `wall`), switched via CSS variables without rebuilding CSS. | `6.0 KB` |
| **Iconography** | **Subsetted SVG `<symbol>` sprite** (`icons.svg`, ~64 symbols) | **Private builds (`-tags fapro`):** Licensed Font Awesome Pro 7.3.1 (**Sharp Light** `fa-sharp fa-light` across navigation, status, and chrome). **Public/upstream builds (default):** Font Awesome Free 7.x fallback sprite emitting identical symbol IDs. Zero runtime FA JS or webfonts. | `13.8 KB` (Pro) / `11.8 KB` (Free) |
| **Typography** | Self-hosted WOFF2 Latin subsets (**IBM Plex Sans** + **IBM Plex Mono**) | Tabular numerals (`tabular-nums`), locally embedded WOFF2 fonts (`IBMPlexSans-Regular/SemiBold/Bold.woff2`, `IBMPlexMono-Regular/Bold.woff2`). Optional `-tags systemfonts` build omits WOFF2 files and drops initial payload from `139.8 KB` to `69.8 KB`. | `46.0 KB` + `24.0 KB` *(optional)* |
| **PWA / Offline Shell** | Web App Manifest + minimal Service Worker (`sw.js`) | Caches immutable hashed assets only (never data or SSE); network-first HTML with "router unreachable" fallback. | `3.0 KB` + `1.6 KB` *(lazy)* |

### HTTP / SSE Transport Details
- **HTTP/2 over TLS** (avoids HTTP/1.1's 6-connection-per-host SSE limit; supports up to 8 streams/session, 32 globally).
- **No compression on `text/event-stream`** (prevents buffering/flush stalls; Overview SSE delta at 5 Hz fast loop + 1 Hz slow piggyback is ~5–10 KB/s). Static assets serve pre-compressed `.br` / `.gz` with zero router CPU cost.
- **Read-only route enforcement:** Declarative HTTP mux allowing only `GET`/`HEAD` on the entire data plane (`POST` exists strictly on `/login` and `/logout` with CSRF tokens).
- **Strict CSP:** `default-src 'none'; script-src 'self' 'unsafe-eval'; style-src 'self'; ...` (`'unsafe-eval'` is required by Datastar v1 expression evaluation and mitigated by a mandatory build-time template lint banning any dynamic `data-*` attribute).

---

## 5. Static Binary (`CGO_ENABLED=0`) Engineering Trade-Offs

The Design Brief specifies `CGO_ENABLED=0` static cross-compilation (Brief §13.2) alongside two Linux features that normally pull in C shared libraries (`cgo`):

| Feature | C / `cgo` Library | Impact on Single Static Binary | Recommended Path for `vyview` |
| --- | --- | --- | --- |
| **journald streaming** | `coreos/go-systemd/sdjournal` (`libsystemd.so`) | Breaks `CGO_ENABLED=0` and couples to host `glibc`/`libsystemd` ABI across VyOS upgrades. | **Use `journalctl -o json -f`** (one long-lived allowlisted child process, ~0.05 ms/event; already endorsed in Brief §9.3 & §13.2). Keeps binary 100% static. |
| **VyOS PAM login** | `libpam.so` | Direct `pam_authenticate` requires `cgo` and host `libpam`. | **Evaluate early in MVP v0.1:** (1) Pure-Go auth first (`argon2id` local users, mTLS client certs from VyOS PKI, forward-auth header) + optional tiny allowlisted OS PAM helper (or `/sbin/unix_chkpwd` / dedicated PAM subprocess) to keep `vyview` `CGO_ENABLED=0`, **or** (2) build with an `arm64` Debian/VyOS cross-toolchain if native PAM in-process is strictly required. |

---

## 6. Build-Time & CI Toolchain (Never Installed on the Router)

Everything below runs exclusively on the developer machine or CI runner:

- **Go toolchain** (pinned version) — `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w'` (plus `amd64`); optional `-tags fapro` for private builds with licensed FA Pro 7.3.1 icons; native Go fuzzing (`go test -fuzz`) and `govulncheck`.
- **`templ` CLI** *(if `templ` is selected over `html/template`)* — generates typed Go code at build time; zero runtime dependency.
- **Node.js (CI-only)** — runs:
  1. **UnoCSS CLI** → compiles `templates/**/*.templ` & `assets/js/**/*.js` into single `assets/app.css`.
  2. **Icon sprite subsetter** → reads `icons.manifest`, validates every icon name against pinned Font Awesome `7.3.1` (from private By-Icon Kit / private npm registry token in CI, or Free 7.x fallback for public builds), tags duotone `<path>` elements with `.fa-primary` / `.fa-secondary`, emits `icons.svg` and `icons.provenance.json`.
  3. **Font Awesome Pro leak guard** → CI gate failing any public build containing Pro SVG path hashes, `@fortawesome/pro-*` references, or `-tags fapro`.
  4. **ECharts modular bundler** → builds minimal Sankey + Heatmap + Graph bundle.
  5. **Asset pipeline** → content hashing + Brotli/Gzip pre-compression + `< 150 KB` payload budget gate.
- **Playwright + `axe-core`** — visual regression across pages × themes × breakpoints, keyboard-only journeys, WCAG 2.2 AA compliance (0 serious/critical violations), and LAN first-paint tracing.
- **Packaging & SBOM** — `.deb` builder (binary + systemd units + default config for custom VyOS image builds, Option A) and CycloneDX SBOM + SHA-256 checksums.

---

## 7. Summary: What Runs on the Router vs. What Does Not

| Runs ON the Router | NEVER Runs on the Router |
| --- | --- |
| `/usr/bin/vyview` (single static Go binary, run as `vyview-web` + `vyview-probe`) | Node.js / npm / UnoCSS CLI / icon subsetter |
| `/var/lib/vyview/vyview.ring` (single pre-allocated multi-ring data file) | SQLite / PostgreSQL / Prometheus / VictoriaMetrics / Grafana |
| `/run/vyview/` (`tmpfs`: live `5 s` hot ring, `probe.sock`, rendered `vyview.toml`) | Python web framework (Flask/FastAPI) or custom Python daemon |
| Existing VyOS/Linux subsystems (kernel netlink, `/proc`, `/sys`, FRR `vtysh`, Kea socket, strongSwan VICI, `journalctl`, `chronyc`, `lldpcli`, `rec_control`, `systemd` D-Bus) | VyOS HTTP API key (not needed in on-box mode) |
| Optional hardware dataplane (`ask2ctl` / DPAA1 sysfs, Suricata stats) | Any shell (`/bin/sh`), CDN, external web-font service, or outbound telemetry |
