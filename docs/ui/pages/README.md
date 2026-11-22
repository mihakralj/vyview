# Page Specifications

> Standardized specifications for every operational page in vyview. Each spec defines the primary operator question, required layout and views, data sources and cadences, and UX/consistency constraints.

---

## Common Page States & Error Rules

Every page inherits these baseline behaviors:
- **Last-known rendering:** Always render the last valid snapshot; show loading skeletons **only** when a source has never been collected.
- **Empty states:** Treat unconfigured features as informational and include a copyable VyOS configuration command hint.
- **Unsupported / absent modules:** Hide hardware- or module-specific sections (e.g., ASK2, VPP, MRP, Tailscale) from navigation when absent, while showing a clear explanation if visited via direct URL.
- **Stale detection:** Mark a widget stale automatically when data age exceeds `2 × cadence`.
- **Source & partial failures:** Retain dimmed last-good data, display the error message and last success timestamp, and explicitly name which portion failed when a page combines multiple collectors.
- **Parser hygiene:** Prefer kernel netlink, `/proc`, `/sys`, and structured JSON sockets; cover text fallbacks with release-specific parser mappings and golden tests.

*(For original brief field tables, consult [brief-extracted.txt](../../reference/brief-extracted.txt) and the [Design Brief PDF](../../reference/VyOS-ViewOnly-Dashboard-Design-Brief.pdf).)*

---

## Page Index

1. [Overview](01-overview.md) — *"Is everything OK, and what changed?"*
2. [Interfaces List](02-interfaces-list.md) — *"Are links up, at what speed, and how are they structured?"*
3. [Interface Detail](03-interface-detail.md) — Deep per-port telemetry (Overview, Counters, Optics, Addresses, Queues, Config)
4. [Traffic & Clients](04-traffic-clients.md) — *"Who is using bandwidth?"* & *"What devices are visible to my router?"*
5. [Routing](05-routing.md) — Per-VRF RIB/FIB, longest-prefix lookup, BGP, OSPF, BFD, and policy routing
6. [Firewall & NAT](06-firewall-nat.md) — Zone policy matrix, dense rule table, **Firewall Security Check**, NAT, and flowtables
7. [VPN](07-vpn.md) — WireGuard (`wgfp_rx`), IPsec (VICI + `xfrm` + CAAM), OpenVPN, and optional overlay VPNs
8. [Services](08-services.md) — Podman containers, control-socket health, Kea DHCP, PowerDNS, chrony/PTP, SSH, LLDP, DDNS, SNMP
9. [QoS](09-qos.md) — Ingress & egress CAKE/FQ-CoDel shapers, all tins, ECN vs. drops, offload bypass, bufferbloat grade
10. [Dataplane / Acceleration](10-dataplane.md) — Unified capability-detected acceleration tree (`ASK2`, `VPP`, `Flowtable`, `AF_XDP`, `CAAM`, `QAT`)
11. [System](11-system.md) — CPU heatmap, PSI, TMP431 thermals, translated eMMC wear, boot images, `softnet_stat`, and self-health
12. [Configuration](12-configuration.md) — Collapsible config tree, single `set` block, red/green diff, audited commit history, CLI cheat sheets
13. [Logs & Events](13-logs-events.md) — Severity-stacked histogram, severity/unit filter chips, journal tail, curated 30 d event timeline
14. [Health & Alerts](14-health-alerts.md) — Engaged vs. past alerts, operational checks, and **Read-Only Security Check**
15. [Diagnostics](15-diagnostics.md) — Step-by-step WAN connectivity chain and scheduled probe trends
16. [Topology](16-topology.md) — Layered L2/L3 network map with LLDP nodes, zone shading, speed/load colors, and MRP rings
17. [Settings](17-settings.md) — Appearance, units, timezone/clock, thresholds, retention summary, `/metrics`, and MCP status
18. [Redundancy](18-redundancy.md) — *"If something fails, what takes over?"* (VRRP, `conntrack-sync`, WAN failover, MRP rings)
