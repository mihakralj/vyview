# vyview Monitoring Gap Analysis & Verification Matrix

> Audits the external monitoring benchmark brief (comparing **ntopng**, **btop++**, **Netdata**, **OPNsense Insight**, **OpenWrt LuCI**, and **MikroTik Torch**) against the actual `vyview` `docs/` and `plans/` specifications for the **NXP LS1046A "Mono Gateway"** (VyOS 1.5/rolling, DPAA1/ASK2).

---

## 1. Executive Verdict

Roughly **75% of the benchmark brief's P0/P1 items were already covered** in `vyview`'s plans (`plans/` + `docs/`). However, the benchmark surfaced **six genuine architectural and collector gaps** that are directly relevant to the LS1046A VyOS hardware—most importantly the **Two-Plane Hardware-Offload Accounting Model** and **VRRP HA monitoring**—all of which have now been integrated into `docs/TECH-STACK.md`, `docs/DESIGN-BRIEF-ANALYSIS.md`, and `plans/`.

### Architectural Spine: The Two-Plane Offload Model
Every general-purpose tool (ntopng, Netdata, Insight, Torch) counts packets at a CPU-visible tap (`libpcap`, `ng_netflow`, `conntrack`, or `netdev`). On LS1046A with ASK2/DPAA1, offloaded flows travel FMan-to-FMan and bypass the CPU entirely. `vyview` therefore adopts a **Two-Plane Model** across Interfaces, Traffic & Clients, Conntrack, and Dataplane (ASK2):

- **Plane A (Hardware / Silicon):** FMan MAC/port `ethtool -S` counters, ASK2 hardware flow table (`fe_flow` debugfs / `ask2ctl`), QMan/BMan pool & FQ congestion, CAAM crypto job counters, and WireGuard fast-path RX (`wgfp_rx`).
- **Plane B (Kernel / CPU):** `rtnetlink` `IFLA_STATS64` (`netdev`), `CTNETLINK` (`conntrack` + `[OFFLOAD]`/`[HW_OFFLOAD]` bits), `nf_flowtable`, `/proc/net/softnet_stat`, `/proc/net/xfrm_stat`.
- **Derived Reconciler:** Joins Plane A (`fe_flow` tuples & FMan port bytes) with Plane B (`conntrack` tuples & `netdev` bytes) so **`offload_ratio`** (`offloaded vs. slow-path` per port and per flow) and **source-split drop counters** (`netdev` vs `ethtool`/FMan FIFO vs `softnet_stat` vs BMan depletion) become first-class P0 metrics.

---

## 2. Verified Mapping Against `vyview` Docs & Plans

Status legend:
- **PLANNED** — Already covered in the original brief/plans.
- **ADDED** — Genuine gap from the benchmark brief, now added to `docs/` and `plans/`.
- **DEFERRED / OUT OF SCOPE** — Intentionally excluded from v1 (or on-box router runtime) to protect ARM64 CPU, RAM (`< 40 MB`), and eMMC budgets.

### 2.1 Data Sources

| Item | Pri | Status | Where Specified & Notes |
| --- | --- | --- | --- |
| Netdev counters (`IFLA_STATS64`, `/sys/class/net/*/statistics`) | P0 | **PLANNED** | `docs/TECH-STACK.md` §2.1 (`200 ms` fast loop, monotonic Δt), `plans/02-pages/03-interface-detail.md` |
| `ethtool -S` (`ETHTOOL_GSTATS`) incl. DPAA/FMan MAC & RX FIFO overflow counters | P0 | **PLANNED + ENHANCED** | `docs/TECH-STACK.md` §2.1, `plans/02-pages/03-interface-detail.md` — enhanced to explicitly split drops by source (`netdev` vs `ethtool`/FMan MAC FIFO vs `softnet` vs BMan) |
| **ASK2 hardware flow table (`fe_flow` debugfs / `ask2ctl`) + offload reconciler** | P0 | **ADDED (P0)** | `docs/TECH-STACK.md` §2.1/§2.3, `plans/02-pages/10-dataplane.md`, `plans/02-pages/04-traffic-clients.md` — elevated from generic TBD to explicit Two-Plane Reconciler (`fe_flow` + `conntrack` join; debugfs reader lives in `vyview-probe`) |
| QMan/BMan/FQ congestion & buffer-pool depletion | P1 | **PLANNED** | `plans/02-pages/10-dataplane.md` (BMan `< 30%` warn / `< 15%` crit), `docs/TECH-STACK.md` §2.3 |
| **CAAM crypto job counters** (`/proc/crypto`, CAAM debugfs vs `/proc/net/xfrm_stat`) | P1 | **ADDED** | `docs/TECH-STACK.md` §2.1, `plans/02-pages/07-vpn.md`, `plans/02-pages/10-dataplane.md` — verifies IPsec hardware offload vs software fallback |
| Conntrack stats & flows (`CTNETLINK`, `/proc/net/stat/nf_conntrack`) | P0 | **PLANNED + ENHANCED** | `docs/TECH-STACK.md` §2.1, `plans/02-pages/04-traffic-clients.md` — enhanced to prefer `CTNETLINK` `NEW`/`DESTROY` events for churn/accounting and reserve full table dumps for on-demand view |
| `nf_flowtable` offloaded entries (`[OFFLOAD]` / `[HW_OFFLOAD]` bits) | P1 | **PLANNED** | `plans/02-pages/06-firewall-nat.md`, `docs/TECH-STACK.md` §2.1 |
| WireGuard peers (`wgctrl`) + **fast-path RX (`wgfp_rx`)** | P0 | **PLANNED + ENHANCED** | `plans/02-pages/07-vpn.md`, `docs/TECH-STACK.md` §2.1 — added explicit `wgfp_rx` fast-path counter reconciliation |
| IPsec SAs (VICI + **`xfrm` netlink & `/proc/net/xfrm_stat`**) | P1 | **PLANNED + ENHANCED** | `plans/02-pages/07-vpn.md`, `docs/TECH-STACK.md` §2.1/§2.2 — added `/proc/net/xfrm_stat` replay/auth error counters alongside VICI `list-sas` |
| **VRRP HA state (`keepalived` D-Bus / event-driven state timeline)** | P0 | **ADDED** | `docs/TECH-STACK.md` §2.1/§2.2, `plans/02-pages/08-services.md`, `plans/02-pages/14-health-alerts.md` — zero-polling D-Bus / netlink event-driven master/backup timeline & flap alert |
| FRR (BGP/OSPF/BFD/RIB) | P1 | **PLANNED** | `docs/TECH-STACK.md` §2.3 (`vtysh` batched JSON at 5 s BFD / 30 s BGP/OSPF) |
| `hwmon`/thermal (TMP431 CPU/PCB/SFP+) & fan RPM | P0 | **PLANNED** | `docs/TECH-STACK.md` §2.1, `plans/02-pages/11-system.md` |
| SFP+ DOM (`ethtool -m` Rx/Tx power, temp) | P1 | **PLANNED** | `docs/TECH-STACK.md` §2.1 (slow 300 s I2C cadence), `plans/02-pages/03-interface-detail.md` |
| **`/proc/net/softnet_stat`** + per-CPU IRQ/NAPI (`/proc/interrupts`, `/proc/softirqs`) | P1 | **PLANNED + ENHANCED** | `docs/TECH-STACK.md` §2.1, `plans/02-pages/11-system.md` — explicitly added `/proc/net/softnet_stat` backlog/time-squeeze/RPS drops alongside `/proc/softirqs` and `/proc/interrupts` |
| **AF_XDP / XDP & VPP stats segment** *(optional overlay)* | P1 | **ADDED (Optional)** | `plans/02-pages/10-dataplane.md` — capability-detected optional adapter when VPP/AF_XDP owns NIC queues (hidden when absent) |
| eMMC wear (`life_time`, `pre_eol_info`) | P1 | **PLANNED** | `docs/TECH-STACK.md` §2.1, `docs/STORAGE-FORMAT.md` §6, `plans/02-pages/11-system.md` |
| DHCP leases (Kea) & neighbor table (`RTM_GETNEIGH`) | P1 | **PLANNED** | `docs/TECH-STACK.md` §2.1/§2.2, `plans/02-pages/04-traffic-clients.md` |
| Firewall rule counters (`nft` netlink) + dead-rule detection | P1 | **PLANNED + ENHANCED** | `plans/02-pages/06-firewall-nat.md` — added "never hit in 30 d" filter alongside hide 0-hit & rule hit sparklines |
| sFlow/NetFlow ingestion from external routers | P2 | **OUT OF SCOPE** | Single-router on-box focus; export only |
| On-box nDPI L7 packet inspection | P2 | **OUT OF SCOPE** | Too CPU-heavy on LS1046A and blind to hardware-offloaded flows |

### 2.2 Collectors, Aggregators & Dashboards

| Item | Pri | Status | Where Specified & Notes |
| --- | --- | --- | --- |
| Single in-process Go collector with multi-tier ticks (no per-tick forks) | P0 | **PLANNED** | `docs/TECH-STACK.md` §2 (`200 ms` fast loop, `1 s`–`300 s` medium/slow) |
| **Event-driven netlink/D-Bus collectors** (`rtnetlink` link/addr/neigh multicast, `ctnetlink` `NEW`/`DESTROY`, `keepalived` VRRP D-Bus) | P0 | **ADDED** | `docs/TECH-STACK.md` §2.1, `plans/05-architecture.md` — gives exact sub-second flap timestamps at lower CPU than polling |
| **Two-Plane Offload Reconciler** (`fe_flow` ↔ `conntrack` + FMan ↔ `netdev`) | P0 | **ADDED** | `docs/TECH-STACK.md` §2.1, `plans/05-architecture.md`, `plans/02-pages/10-dataplane.md` |
| Self-telemetry (collector duration, dropped ticks, `/proc/self/io` write bytes, RSS) | P0 | **PLANNED** | `plans/02-pages/11-system.md`, `docs/STORAGE-FORMAT.md` §6 |
| Prometheus/OpenMetrics `/metrics` export | P1 | **PLANNED** | `docs/TECH-STACK.md` §3, `plans/appendices.md` |
| **Space-saving / heavy-hitter Top-K sketch** for bounded per-host/per-flow tracking | P0 | **ADDED** | `docs/STORAGE-FORMAT.md` §3.3, `plans/02-pages/04-traffic-clients.md` — bounds RAM strictly under 8k–50k concurrent sessions |
| Rate math with monotonic Δt, 32/64-bit wrap & ASK2 driver-reload reset handling | P0 | **PLANNED + ENHANCED** | `docs/STORAGE-FORMAT.md` §3.2 — explicitly includes debugfs/driver-reload counter reset handling |
| Multi-tier RAM + single-file fixed rollups (`vyview.ring`) | P0 | **PLANNED** | `docs/STORAGE-FORMAT.md` (`200 ms` → `5 s` → `60 s` → `5 m` → `1 h`, plus Compact profile) |
| On-demand time-boxed "Torch" capture (`AF_PACKET` + BPF) | P1 | **DEFERRED (v2)** | Requires `CAP_NET_RAW`, cannot see ASK2 offloaded flows, and adds packet-capture security surface; kept out of v1 (which already has conntrack + `fe_flow` live top talkers) |
| Host behavioural score / ASN-GeoIP MMDB | P2 | **DEFERRED (v2+)** | Low value without L7 DPI; optional lazy MMDB deferred past v0.4 |
| External TSDB on router (InfluxDB / ClickHouse / Grafana) | Out | **REJECTED (Red Flag)** | Explicitly banned in `docs/TECH-STACK.md` §7 |

---

## 3. Red-Flag Audit Result

All red flags cited in the benchmark brief are **already avoided by design** in `vyview`:
- No on-box InfluxDB, ClickHouse, Prometheus TSDB, SQLite WAL, or Grafana.
- No always-on `libpcap` / nDPI inspection.
- Bounded cardinality caps (`760` fixed series max, top-50 clients + `other`, top-120 firewall rules) enforced by the fixed-size `vyview.ring` binary layout.
- Self-telemetry (`/proc/self/io`, collector tick duration, eMMC `life_time` / `pre_eol_info`) verifies the `< 3%` CPU, `< 40 MB` RSS, and `< 50 MB/day` flash-write budgets continuously.
