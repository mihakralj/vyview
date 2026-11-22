# Feedback Log — @Haico (Nov 21, 2026)

> Review and incorporation log for `@Haico`'s 25-point OT/prosumer review of the design brief and `v0.1.0-demo` running on a Mono Gateway (`VyOS 2026.11.21-0555-rolling`).

---

## 1. Core UX Principles Adopted

Promoted into [Product Overview §3](../architecture/overview.md#3-five-binding-ux-design-principles) and applied across every page spec:
1. **Every page answers one question** — Replace generic "Records" scaffolding with purpose-built operational views.
2. **Explain, don't just show** — State *why* a warning exists, *why* a flow is not hardware-offloaded, and *what to do next*.
3. **One truth everywhere** — Eliminate 2×–3× dataset duplication on a page and reconcile numbers across cross-linked views.
4. **From glance to depth** — Lead with color + icon shape + plain-language health sentence; put deep telemetry underneath.
5. **No magic numbers** — Pair raw hex codes, register fields, and thresholds with plain-language meaning.

---

## 2. Point-by-Point Decisions & Traceability

| # | Topic | Decision | Canonical Spec Updated |
| --- | --- | --- | --- |
| **1** | **Opt-in MCP integration (`commit-confirm`)** | **ADOPTED** — Web GUI stays strictly read-only; opt-in MCP interface supports read queries, gated diagnostics, and audit-logged writes exclusively via VyOS `set → commit-confirm → save`. | [Product Overview](../architecture/overview.md), [Security & MCP](../architecture/security-and-mcp.md) |
| **2** | **Plain-language workflows & Read-Only Security Check** | **ADOPTED** — Added 7 core question paths (`≤ 3` clicks / `1` MCP prompt), step-by-step WAN failure chain, asset inventory, and Read-Only Security Check modules. | [Product Overview](../architecture/overview.md), [Firewall & NAT](../ui/pages/06-firewall-nat.md), [Health & Alerts](../ui/pages/14-health-alerts.md), [Diagnostics](../ui/pages/15-diagnostics.md) |
| **3** | **Unified Dataplane / Acceleration + ASK2 Netlink (`ynl`) + AF_XDP / TSN** | **ADOPTED (CRITICAL CORRECTION)** — Unified capability-detected tree (`Summary`, `Forwarding: ASK2 / VPP / Flowtable / AF_XDP`, `Crypto: CAAM / QAT`, `Other`); prefer ASK2 Generic Netlink (`ynl --family ask` / `show offload flow`) over `fe_flow` debugfs; added AF_XDP ZC and OT/TSN L2/IP protocols. | [Dataplane / Acceleration](../ui/pages/10-dataplane.md), [System Architecture](../architecture/system-architecture.md) |
| **4** | **OT / IT-OT convergence (Persona P6)** | **ADOPTED** — Added OT engineer persona covering IT/OT zone separation, change traceability, L2/TSN protocol visibility, and determinism. | [Product Overview](../architecture/overview.md), [Traffic & Clients](../ui/pages/04-traffic-clients.md), [Topology](../ui/pages/16-topology.md) |
| **5** | **Lowering the barrier for other vendors** | **ADOPTED** — FortiGate/OPNsense-style policy table, multi-vendor CLI cheat sheets (Cisco, Juniper, Fortinet, MikroTik, Hirschmann, Siemens SCALANCE), MCP migration diffs. | [Firewall & NAT](../ui/pages/06-firewall-nat.md), [Configuration](../ui/pages/12-configuration.md) |
| **6** | **Prometheus `/metrics` as recommended KISS path** | **ADOPTED** — Export all telemetry plus vyview's own health checks and alerts on `/metrics` so Prometheus + Alertmanager (optionally as a VyOS Podman container) handle notifications. | [System Architecture](../architecture/system-architecture.md), [Health & Alerts](../ui/pages/14-health-alerts.md) |
| **7** | **MRP (IEC 62439-2) OT ring display** | **ADOPTED** — Read-only ring status, MRM/MRC role, ring ports, blocked link, and ring events once supported in VyOS. | [Redundancy](../ui/pages/18-redundancy.md), [Topology](../ui/pages/16-topology.md), [Interfaces List](../ui/pages/02-interfaces-list.md) |
| **8** | **OT & MCP success metrics** | **ADOPTED** — Added MRP visibility (`< 2 s`, `100%` capture), real-time latency no-harm (`< X µs`), PTP (`1 s`), `x86`/`LS1046A` CI parity, `≥ 95%` OT asset discovery, MCP 1-question & write-safety metrics. | [Product Overview](../architecture/overview.md), [Performance & Testing](../architecture/performance-and-testing.md) |
| **9** | **Shell & navigation polish** | **ADOPTED** — Top-bar health pill, live/pause, theme toggle, user menu; CLI-aware `Ctrl-K` palette; HTTPS default on LAN (Clipboard API); full breadcrumbs; collapsible rail tree. | [Information Architecture](../ui/information-architecture.md) |
| **10** | **Interfaces list (3-layer view)** | **ADOPTED** — Front-panel physical port strip + sortable core table (status dot+text, IPv4+IPv6, master/members, labelled sparks, period deltas) + logical bridge/VLAN/MRP tree. | [Interfaces List](../ui/pages/02-interfaces-list.md) |
| **11** | **Best-in-class GUI goal** | **ADOPTED** — Retire generic Records scaffolding across all 18 page specs. | [Product Overview](../architecture/overview.md), [Page Specifications](../ui/pages/README.md) |
| **12** | **Interface detail fixes** | **ADOPTED** — Reconciled `Errors` tile with `rx_fcs_err`, clickable flaps, LLDP card, mirrored RX/TX chart with axes, SFP-only Optics with warning+alarm limits & trends, desktop drawer with real IDs and no Copy on `—`, separated Addresses vs. Gateway/Neighbors, SW+HW Queues with offload bypass, full Config subtree + drift. | [Interface Detail](../ui/pages/03-interface-detail.md) |
| **13** | **Traffic & Clients** | **ADOPTED** — Single rich table (no 3× duplication), explicit `"top N of total"`, connection type separated from status, stacked top-6+other area chart, RX+TX balance, OT asset metadata, true WAN↔segment + inter-VLAN Sankey. | [Traffic & Clients](../ui/pages/04-traffic-clients.md) |
| **14** | **Dedicated Redundancy section** | **ADOPTED** — Top-level **Redundancy** hub (`VRRP`, `conntrack-sync`, `WAN failover`, `MRP`). | [Redundancy](../ui/pages/18-redundancy.md), [Information Architecture](../ui/information-architecture.md) |
| **15** | **Firewall** | **ADOPTED** — Zone matrix above/beside single dense grouped rule table, reconciled default-policy drops, hit bars, last hit, 30 d dead-rule marker, compact icon columns, density toggle. | [Firewall & NAT](../ui/pages/06-firewall-nat.md) |
| **16** | **Conntrack & NAT** | **ADOPTED** — Dense filterable table, inline NAT column, DNS/client resolution, offload `%` + explicit path, matching firewall rule link, `insert_failed` + `early_drop`, *"why not offloaded"*, DNS-bypass signal. | [Traffic & Clients](../ui/pages/04-traffic-clients.md) |
| **17** | **QoS** | **ADOPTED** — Both ingress & egress shapers, all tins (warn Voice/Video, not Bulk), totals unified with Interface Queues, ECN marks + drops, multi-tin delay chart, offload-bypass share. | [QoS](../ui/pages/09-qos.md) |
| **18** | **Topology** | **ADOPTED** — Layered L2/L3 map, BGP badge on WAN link, dashed VPNs, LLDP switches/APs, collapsible client groups, firewall zone shading, negotiated speed & load colors, MRP rings. | [Topology](../ui/pages/16-topology.md) |
| **19** | **System → Storage & images** | **ADOPTED** — Translated eMMC Type A/B & Pre-EOL (`0x02`/`0x03` red), full installed image list, per-partition usage + room-for-image check. | [System](../ui/pages/11-system.md) |
| **20** | **Dataplane / Acceleration** | **ADOPTED** — Single split HW/SW bar, disambiguated offload % (flows/packets/bytes), stacked HW/SW chart, `%` pool bars, plain-language Suricata shunt, *"why flows aren't offloaded"*. | [Dataplane / Acceleration](../ui/pages/10-dataplane.md) |
| **21** | **Configuration** | **ADOPTED** — Collapsible tree linked to status pages, single copyable `set` block, red/green two-revision diff, audited commit history (user/source/comment/`commit-confirm`, absolute TZ), no Copy on `[REDACTED]`. | [Configuration](../ui/pages/12-configuration.md) |
| **22** | **Logs & Events** | **ADOPTED** — Remove duplicate timeline, severity-stacked histogram with axes, separated Severity vs. State, full TZ timestamps, explicit pool/threshold context, object links + next steps, filter chips. | [Logs & Events](../ui/pages/13-logs-events.md) |
| **23** | **Diagnostics** | **ADOPTED** — Step-by-step WAN failure chain, explicit probe targets/resolvers/thresholds/units, historical trend charts, on-demand tests via opt-in MCP. | [Diagnostics](../ui/pages/15-diagnostics.md) |
| **24** | **Overview** | **ADOPTED** — Top-bar health pill + sentence, unambiguous wording (`Lease 21 h remaining`, sensor name, offload unit), Download/Upload labels + local-traffic note, explicit Standby/Down on `pppoe0`, verified Wi-Fi source, linked talkers/events, no rail jump links. | [Overview](../ui/pages/01-overview.md) |
| **25** | **Settings** | **ADOPTED** — Separate labelled WCAG radio/segmented groups, units, TZ/24 h clock, default range, plain-language warning thresholds, retention summary, `/metrics` & MCP status. | [Settings](../ui/pages/17-settings.md) |
