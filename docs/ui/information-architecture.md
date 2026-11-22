# Information Architecture

> Defines vyview's application shell, navigation hierarchy, top-bar controls, detail drawers, section inventory, and URL/state contract.

---

## 1. Shell & Navigation Model

### Left Activity Rail
- Organizes destinations into four clear groups — **MONITOR**, **NETWORK**, **PLATFORM**, and **INSPECT** — with **Settings** pinned to the footer.
- **Desktop width:** `232 px`, collapsible to a `64 px` icon rail (keyboard shortcut `[`). Sections expand into collapsible sub-trees showing the active entity or tab.
- **Mobile (`< 640 px`):** Five-destination bottom navigation bar plus a **More** sheet.
- **No ad-hoc jump lists in the rail:** Never place `"On this page"` anchor links inside the left application rail.

### Top Bar & Contextual Chrome
- **Full hierarchical breadcrumbs:** `Router › Section › Entity › Sub-view` (for example, `demo-router › Interfaces › eth0 › Optics`).
- **Global search / command palette (`Ctrl/Cmd-K`):** Searches across entities, pages, config paths, and recognizes VyOS CLI commands (e.g., `show interfaces`), clearly explaining why nothing matched when empty.
- **Shared time-range picker & Live/Pause toggle:** Synchronized across all charts on the page, displaying a pending-update counter while paused.
- **Top-bar Health pill:** Shows worst active status, alert count, and a plain-language summary linking directly to **Health & Alerts**.
- **Theme toggle & User menu:** Quick theme switching and session controls.
- **Page header actions:** Sub-view tabs, copy permalink, pin view, CSV/JSON/PNG export, and copyable source commands.

### Detail Drawers
- Entity details open in a **right-hand drawer on desktop** (full-screen sheet on phones) with trapped keyboard focus, `Escape` to close, and real entity URL IDs (e.g., `?sel=rx_fcs_err`, never opaque shorthands).

---

## 2. Section Inventory

| Group | Section | Primary Operational Views |
| --- | --- | --- |
| **MONITOR** | **[Overview](pages/01-overview.md)** | Plain-language health summary, WAN card (failure chain), hero throughput, 7 KPIs, linked top talkers & dated events; kiosk mode. |
| **MONITOR** | **[Health & Alerts](pages/14-health-alerts.md)** | Active vs. past alerts, built-in operational checks, **Read-Only Security Check** (*"Am I protected well enough?"*). |
| **NETWORK** | **[Interfaces](pages/02-interfaces-list.md)** | 3-layer view (physical port panel + core table + logical bridge/VLAN/MRP tree) & type-specific [Interface Detail](pages/03-interface-detail.md) (Overview, Counters, Optics on SFP only, Addresses, Queues, Config). |
| **NETWORK** | **[Traffic & Clients](pages/04-traffic-clients.md)** | Stacked top-6 + other chart, client/OT asset inventory, protocol mix, WAN↔segment & inter-VLAN Sankey, virtualized Conntrack & NAT table. |
| **NETWORK** | **[Routing](pages/05-routing.md)** | VRF selector & per-VRF RIB/FIB v4/v6 tabs (`default` first), longest-prefix lookup, static, BGP, OSPF, BFD, policy routing, churn. |
| **NETWORK** | **[Firewall & NAT](pages/06-firewall-nat.md)** | Zone policy matrix above/beside dense FortiGate/OPNsense-style grouped rule table, **Firewall Security Check** (*"Is my firewall safe?"*), groups, NAT/NAT66, flowtables. |
| **NETWORK** | **[VPN](pages/07-vpn.md)** | Unified VPN hub: WireGuard (with `wgfp_rx`), IPsec (VICI + `xfrm` + CAAM), OpenVPN, optional Tailscale/ZeroTier. |
| **NETWORK** | **[QoS](pages/09-qos.md)** | Ingress & egress CAKE/FQ-CoDel shapers, all configured tins, ECN marks vs. drops, offload-bypass share, bufferbloat grade. |
| **NETWORK** | **[Topology](pages/16-topology.md)** | Layered L2/L3 auto-drawn map (WAN, router, dashed VPNs, bridges/VLANs, LLDP switches/APs, collapsible client groups, zone shading, negotiated speed & load colors, MRP rings). |
| **NETWORK** | **[Redundancy](pages/18-redundancy.md)** | Capability-detected failover hub: Router (VRRP), State sync (`conntrack-sync`), WAN failover, Rings (MRP). |
| **PLATFORM** | **[Services](pages/08-services.md)** | **Containers (Podman)**, Kea DHCP, PowerDNS, chrony & PTP (`pmc`/`phc2sys`), SSH sessions, LLDP, mDNS, RA, DDNS, UPnP, SNMP — with per-daemon control-socket health. |
| **PLATFORM** | **[Dataplane / Acceleration](pages/10-dataplane.md)** | Capability-detected unified tree: Summary (always first), Forwarding (`ASK2`, `VPP`, `Flowtable`, `AF_XDP`), Crypto (`CAAM`, `QAT`), Other. |
| **PLATFORM** | **[System](pages/11-system.md)** | Per-core CPU heatmap, load/PSI, memory, TMP431 sensors, Storage & images (translated eMMC Type A/B + Pre-EOL, all boot images, partition space), top-10 processes, systemd units, kernel/modules, IRQ/`softnet_stat`, vyview self-health. |
| **INSPECT** | **[Configuration](pages/12-configuration.md)** | Collapsible config tree (linked to status pages), single copyable `set` block, JSON, red/green two-revision diff, audited commit history (user/CLI-API-MCP/comment/`commit-confirm`), multi-vendor CLI cheat sheets. |
| **INSPECT** | **[Logs & Events](pages/13-logs-events.md)** | Severity-stacked histogram, severity & unit filter chips, journal tail, firewall/auth logs, curated 30 d event timeline. |
| **INSPECT** | **[Diagnostics](pages/15-diagnostics.md)** | Step-by-step WAN connectivity chain (*"Why is my internet not working?"*), scheduled latency/jitter/loss/DNS/traceroute trends & explicit thresholds (on-demand tests via MCP). |
| **FOOTER** | **[Settings](pages/17-settings.md)** | Separate labelled radio/segmented groups: Appearance, Demo scenario (demo only), Units, Timezone & 24 h clock, Default range, Warning thresholds, Retention summary, `/metrics` status, MCP read-only status. |

---

## 3. URL & State Contract

- **URL paths** identify primary entities (see full route map in [Appendices](../reference/appendices.md#route-map)).
- **Query parameters** capture shareable view state: `range`, `from`, `to`, `filter`, `sort`, `tab`, `sel`, `zone`, and `cmp` (previous-period comparison).
- **Browser `localStorage`** stores personal client preferences: theme, density, units (bits/bytes, `°C`/`°F`), timezone/24 h clock, table column choices, and pinned views.
- **HTTPS by default on LAN:** Ensures secure context across LAN clients so browser Clipboard API (`Copy`) works reliably on every page.
