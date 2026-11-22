# Monitoring Gap Analysis & Verification Matrix

> Audits an external monitoring benchmark comparing **ntopng**, **btop++**, **Netdata**, **OPNsense Insight**, **OpenWrt LuCI**, and **MikroTik Torch** against vyview's specifications for the **NXP LS1046A "Mono Gateway"** (VyOS 1.5/rolling, DPAA1/ASK2).

---

## 1. Executive Verdict

Roughly **75% of the benchmark's P0/P1 items were already planned** in vyview. However, the benchmark surfaced **six genuine architectural and collector gaps** directly relevant to LS1046A VyOS hardware — led by the **Two-Plane Hardware-Offload Accounting Model** and **event-driven VRRP HA monitoring** — all of which are now integrated into the canonical specifications.

### Architectural Spine: The Two-Plane Offload Model
Every general-purpose tool (`ntopng`, `Netdata`, `OPNsense Insight`, `MikroTik Torch`) counts packets at a CPU-visible tap (`libpcap`, `ng_netflow`, `conntrack`, or `netdev`). On LS1046A with ASK2/DPAA1, offloaded flows travel FMan-to-FMan and bypass the CPU entirely. vyview reconciles **Plane A (Hardware/Acceleration)** against **Plane B (Kernel/CPU)** in-process so `offload_ratio` (flows, packets, and bytes) and source-split drop counters (`netdev` vs. FMan FIFO vs. `softnet_stat` vs. BMan depletion) become first-class P0 metrics. See [System Architecture §2](../architecture/system-architecture.md#2-two-plane-hardware-offload-model).

---

## 2. Verification Matrix

Status legend: **PLANNED** (already in original brief), **ADDED** (gap adopted into canonical specs), **DEFERRED / OUT OF SCOPE** (intentionally excluded to protect ARM64 CPU, RAM, and eMMC budgets).

### 2.1 Data Sources

| Item | Pri | Status | Canonical Location & Notes |
| --- | --- | --- | --- |
| Netdev counters (`IFLA_STATS64`) | P0 | **PLANNED** | [System Architecture](../architecture/system-architecture.md) (`200 ms` fast loop, monotonic $\Delta t$), [Interface Detail](../ui/pages/03-interface-detail.md) |
| `ethtool -S` (`ETHTOOL_GSTATS`) incl. FMan MAC & RX FIFO overflows | P0 | **PLANNED + ENHANCED** | [Interface Detail](../ui/pages/03-interface-detail.md) — explicitly splits drops by source (`netdev` vs. FMan FIFO vs. `softnet_stat` vs. BMan) |
| **ASK2 hardware flow table + Two-Plane Offload Reconciler** | P0 | **ADDED (P0)** | [System Architecture](../architecture/system-architecture.md), [Dataplane / Acceleration](../ui/pages/10-dataplane.md) — prefers Generic Netlink (`ynl --family ask` / `show offload flow`) over legacy `fe_flow` debugfs |
| QMan/BMan/FQ congestion & buffer-pool depletion | P1 | **PLANNED** | [Dataplane / Acceleration](../ui/pages/10-dataplane.md) (BMan `< 30%` warn / `< 15%` crit) |
| **CAAM / QAT crypto job counters** vs. `/proc/net/xfrm_stat` | P1 | **ADDED** | [VPN](../ui/pages/07-vpn.md), [Dataplane / Acceleration](../ui/pages/10-dataplane.md) — verifies IPsec HW crypto vs. SW fallback |
| Conntrack stats & flows (`CTNETLINK`) | P0 | **PLANNED + ENHANCED** | [Traffic & Clients](../ui/pages/04-traffic-clients.md) — prefers `NEW`/`DESTROY` events for churn; full table dump on demand only |
| `nf_flowtable` offloaded entries (`[OFFLOAD]`/`[HW_OFFLOAD]`) | P1 | **PLANNED** | [Firewall & NAT](../ui/pages/06-firewall-nat.md), [System Architecture](../architecture/system-architecture.md) |
| WireGuard peers (`wgctrl`) + **fast-path RX (`wgfp_rx`)** | P0 | **PLANNED + ENHANCED** | [VPN](../ui/pages/07-vpn.md) — reconciles `wgctrl` with ASK2 `wgfp_rx` |
| IPsec SAs (VICI + **`NETLINK_XFRM` & `/proc/net/xfrm_stat`**) | P1 | **PLANNED + ENHANCED** | [VPN](../ui/pages/07-vpn.md) — adds replay/auth error counters alongside VICI |
| **VRRP HA state (`keepalived` D-Bus)** | P0 | **ADDED** | [Redundancy](../ui/pages/18-redundancy.md), [Health & Alerts](../ui/pages/14-health-alerts.md) — zero-polling event-driven master/backup timeline |
| FRR (BGP/OSPF/BFD/RIB) | P1 | **PLANNED** | [Routing](../ui/pages/05-routing.md), [System Architecture](../architecture/system-architecture.md) (`vtysh` batched JSON) |
| `hwmon` / TMP431 thermal & fan RPM | P0 | **PLANNED** | [System](../ui/pages/11-system.md) |
| SFP+ DOM (`ethtool -m` Rx/Tx power, temp) | P1 | **PLANNED** | [Interface Detail](../ui/pages/03-interface-detail.md) (`300 s` slow I2C poll) |
| **`/proc/net/softnet_stat`** + per-CPU IRQ/NAPI | P1 | **PLANNED + ENHANCED** | [System](../ui/pages/11-system.md) — adds backlog/time-squeeze/RPS drops |
| **AF_XDP / XDP & VPP stats segment** *(optional)* | P1 | **ADDED** | [Dataplane / Acceleration](../ui/pages/10-dataplane.md) — capability-detected, hidden when absent |
| eMMC wear (`life_time`, `pre_eol_info`) | P1 | **PLANNED** | [System](../ui/pages/11-system.md), [Storage Format](../architecture/storage-format.md) — kernel sysfs only (no `mmc-utils`) |
| DHCP leases (Kea) & neighbor table (`RTM_GETNEIGH`) | P1 | **PLANNED** | [Traffic & Clients](../ui/pages/04-traffic-clients.md), [Services](../ui/pages/08-services.md) |
| Firewall rule counters (`nft` netlink) + dead-rule detection | P1 | **PLANNED + ENHANCED** | [Firewall & NAT](../ui/pages/06-firewall-nat.md) — adds *"never hit in 30 d"* filter |
| sFlow/NetFlow ingestion from other routers | P2 | **OUT OF SCOPE** | Single-router on-box focus; export only |
| On-box nDPI L7 packet inspection | P2 | **OUT OF SCOPE** | Too CPU-heavy on LS1046A and blind to hardware-offloaded flows |

### 2.2 Collectors, Aggregators & Dashboards

| Item | Pri | Status | Canonical Location & Notes |
| --- | --- | --- | --- |
| Single in-process Go collector (no per-tick forks) | P0 | **PLANNED** | [System Architecture §3](../architecture/system-architecture.md#3-stage-1--data-gathering-pipeline) |
| **Event-driven netlink / D-Bus collectors** (`rtnetlink`, `ctnetlink`, `keepalived`, `bridge mrp`) | P0 | **ADDED** | [System Architecture §3.1](../architecture/system-architecture.md#31-go-collector-libraries-compiled-into-vyview) |
| **Two-Plane Offload Reconciler** | P0 | **ADDED** | [System Architecture §2](../architecture/system-architecture.md#2-two-plane-hardware-offload-model) |
| Self-telemetry (tick duration, dropped ticks, `/proc/self/io`, RSS) | P0 | **PLANNED** | [System](../ui/pages/11-system.md), [Storage Format §6](../architecture/storage-format.md#6-vyos-persistence-location-upgrades--self-monitoring) |
| Prometheus `/metrics` export (incl. health & alerts) | P1 | **PLANNED + PROMOTED** | [System Architecture §4](../architecture/system-architecture.md#4-stage-2--storage--metrics-export) |
| **Space-Saving / Heavy-Hitter Top-K sketch** | P0 | **ADDED** | [Storage Format §3.4](../architecture/storage-format.md#34-series-catalog--cardinality-caps), [Traffic & Clients](../ui/pages/04-traffic-clients.md) |
| Monotonic $\Delta t$ rate math + wrap/driver-reload resets | P0 | **PLANNED + ENHANCED** | [Storage Format §3.3](../architecture/storage-format.md#33-rate-math-gaps--peak-preservation) |
| Multi-tier RAM + single-file fixed rollups (`vyview.ring`) | P0 | **PLANNED** | [Storage Format](../architecture/storage-format.md) |
| On-demand time-boxed "Torch" (`AF_PACKET` + BPF) | P1 | **DEFERRED** | Requires `CAP_NET_RAW` and cannot see ASK2 offloaded flows |
| Host behavioural score / ASN-GeoIP MMDB | P2 | **DEFERRED** | Optional lazy MMDB deferred past `v0.4` |
| External TSDB on router (InfluxDB / ClickHouse / Grafana) | Out | **REJECTED** | Explicitly banned in [System Architecture §7](../architecture/system-architecture.md#7-on-router-runtime-vs-ci-only-build-toolchain) |

---

## 3. Red-Flag Audit Verdict

All red flags identified in the benchmark are avoided by construction: no on-box InfluxDB/ClickHouse/SQLite WAL/Grafana, no always-on `libpcap`/nDPI, strict cardinality caps (`760` series max, top-50 clients + `other`, top-120 rules), and continuous self-telemetry verifying `< 3%` CPU, `< 40 MB` RSS, and `< 50 MB/day` flash writes.
