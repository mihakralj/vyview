# Feedback Log — @Daegara (Nov 21, 2026)

> Review and incorporation log for `@Daegara`'s three technical reviews (`TECH-STACK.md`, `v0.1.0-demo`, and `STORAGE-FORMAT.md`, tested on Mono GW DK with `VyOS 2026.11.21-0555-rolling`), cross-correlated with `@Haico`'s review ([feedback-haico.md](feedback-haico.md)).

---

## 1. Cross-Reviewer P0 Convergences

Items independently flagged by both `@Daegara` and `@Haico` on real Mono Gateway hardware:

| Convergence | @Daegara | @Haico | Canonical Spec |
| --- | --- | --- | --- |
| **Physical port strip** | Front-panel view of physical ports & parent status | 3-layer Interfaces page with front-panel port strip | [Interfaces List](../ui/pages/02-interfaces-list.md) |
| **eMMC wear wording & sysfs** | Reframe `"eMMC life"` to `"eMMC wear"` / remaining %; clarify no `mmc-utils` | Translate `0x03` → `"20–30% worn"`, show Type A, Type B, Pre-EOL | [System](../ui/pages/11-system.md), [System Architecture](../architecture/system-architecture.md) |
| **Log-level filtering** | Filter by log level (`INFO`, `CRITICAL`) | Severity & unit filter chips + severity-stacked histogram | [Logs & Events](../ui/pages/13-logs-events.md) |
| **Real layered Topology graph** | Demo looks disconnected; weigh graph libs vs. payload | Layered L2/L3 map with LLDP, zones, speed colors | [Topology](../ui/pages/16-topology.md), [System Architecture](../architecture/system-architecture.md) |
| **Firewall rule-order sort & matrix reconciliation** | Default sort by rule evaluation index; Rule 90 missing in matrix | Preserve evaluation order; default-policy & rule counts match Dropped tile | [Firewall & NAT](../ui/pages/06-firewall-nat.md) |

---

## 2. Technical Additions & Engineering Resolutions

### 2.1 System Architecture & Runtime
| Topic | Engineering Resolution | Canonical Spec |
| --- | --- | --- |
| **Local vs. Remote deployment scope** | **ADOPTED:** On-router Linux (`arm64` LS1046A + `x86_64`) is primary; documented explicit Local vs. Remote trade-offs (TLS validity, required `vyview.toml`, API/SSH transport, security). | [System Architecture §8](../architecture/system-architecture.md#8-local-on-router-vs-remote-off-router-deployment-differences) |
| **Collector resilience & load smoothing** | **ADOPTED:** Per-source timeouts, exponential backoff (to `5 min`) + circuit-breaker limits, phase-staggered scheduling (`±10%` jitter + offsets), and hard ICMP/DNS probe rate caps (`≤ 1/target/10 s`, max `8` targets). | [System Architecture §3](../architecture/system-architecture.md#3-stage-1--data-gathering-pipeline) |
| **Control-socket health & Podman containers** | **ADOPTED:** Surface per-daemon socket health (`Kea`, `VICI`, `Podman`, `D-Bus`, `vyview-probe`) in Services & Health; add read-only **Podman containers** view via `podman.sock` (`GET` only, no start/stop/exec). | [System Architecture §3.2](../architecture/system-architecture.md#32-local-control-sockets), [Services](../ui/pages/08-services.md) |
| **`EXT_CSD` without `mmc-utils`** | **CLARIFIED:** Linux 6.18+ `mmc_block` exposes `EXT_CSD` directly in unprivileged sysfs (`/sys/block/mmcblk*/device/{life_time,pre_eol_info}`). Zero external binary needed. | [System Architecture §3.1](../architecture/system-architecture.md#31-go-collector-libraries-compiled-into-vyview), [System](../ui/pages/11-system.md) |
| **No `ps` / `ss` / `w` forks** | **CLARIFIED:** Built directly in Go from `/proc/<pid>`, `NETLINK_SOCK_DIAG`, and `/run/utmp` / `sshd` journal. | [System Architecture §3.3](../architecture/system-architecture.md#33-allowlisted-read-only-executables-no-shell) |
| **VyOS `operator-group` & pubkey-only auth** | **ADOPTED:** Align with `set system login operator-group vyview-ro command-policy allow 'show'` + pubkey-only SSH for remote installs; validate PAM vs. VyOS users in v0.1. | [Deployment & Packaging §2](../architecture/deployment-and-packaging.md#2-proposed-vyos-cli--rbac-sketch-option-c), [Security & MCP §2](../architecture/security-and-mcp.md#2-network-binding-tls--authentication) |
| **Multi-VRF Routing tabs** | **ADOPTED:** Explicit VRF selector & per-VRF RIB/FIB tabs (`default` first + named VRFs). | [Routing](../ui/pages/05-routing.md) |
| **Topology library trade-off** | **EVALUATED:** Server SVG default + modular lazy ECharts (already bundled for Sankey/heatmaps); Cytoscape/D3 recorded as alternatives guarded by the `< 150 KB` initial budget. | [System Architecture §5](../architecture/system-architecture.md#5-stage-3--presentation--streaming-datastar--browser), [Topology](../ui/pages/16-topology.md) |

### 2.2 Storage Format (`vyview.ring`)
| Topic | Engineering Resolution | Canonical Spec |
| --- | --- | --- |
| **Configurable `--data-path` (USB) & startup check** | **ADOPTED:** Allow external USB/custom mount path with startup availability, free-space (`≥ 20%` headroom), and write-latency checks. | [Storage Format §1–2](../architecture/storage-format.md), [Deployment & Packaging](../architecture/deployment-and-packaging.md) |
| **Delayed boot start & persistence flow** | **ADOPTED:** `After=network-online.target vyos-router.service` + configurable `--start-delay`; clarified `.deb` / `vyos-postconfig-bootup.script` persistence until native config node lands. | [Deployment & Packaging §3–4](../architecture/deployment-and-packaging.md) |
| **Single file vs. tick/tock shadow copy** | **EVALUATED & ADOPTED SELECTIVELY:** Rejected per-flush file duplication (10,000× flash wear); explained eMMC FTL wear-leveling + A/B superblocks + per-slot CRC32C; added optional slow/clean-shutdown passive `--snapshot-path`. | [Storage Format §1](../architecture/storage-format.md#1-design-goals--constraints) |
| **Composable Profile × Retention + `> 50%` wear warning** | **ADOPTED:** Composable `--profile` × `max_retention` with upper bounds, graceful `live-only` fallback on missing target, and early `> 50%` eMMC wear advisory (`EXT_CSD >= 0x06`). | [Storage Format §2](../architecture/storage-format.md#2-memory-tiers-vs-persistent-file-regions), [Health & Alerts](../ui/pages/14-health-alerts.md) |
| **Bounded T2+ warming cache & statistical rollups** | **ADOPTED:** Added `≤ 4 MiB` TTL-evicted (`5 min`) in-RAM query cache for on-disk tiers and explicit statistical cross-tier rollup math (`min`, `max`, weighted `avg`, merged `p95`). | [Storage Format §4.1 & §5](../architecture/storage-format.md) |
| **Schema upgrade & expanded self-monitoring** | **ADOPTED:** Fresh `.vN.ring` + read-only prior mount on schema bumps; expose Superblock A & B health, torn-slot count, and saved vs. nominal slots. | [Storage Format §6](../architecture/storage-format.md#6-schema-upgrades--self-monitoring), [System](../ui/pages/11-system.md) |

---

## 3. Demo Fixture Follow-Ups (Next Demo Code Pass)
- Align WireGuard fixtures so both `wg0` and `wg1` appear consistently across Interfaces and VPN (One Truth).
- Ensure Firewall default sort follows numeric rule evaluation order and mirrors Rule 90 cleanly into the `WAN → LAN` zone matrix.
