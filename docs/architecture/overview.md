# Product Overview

> Defines vyview's product vision, boundaries, five UX design principles, target personas, core plain-language question workflows, and measurable success criteria.

---

## 1. Vision & Platform Scope

**vyview** is a single-router, LAN-only, real-time and historical dashboard for [VyOS](https://vyos.io). It provides fast, transparent operational visibility into everything VyOS op-mode and the Linux kernel already know — without requiring SSH, external databases, or cloud services.

- **Web GUI is strictly read-only by construction:** There are no mutating web endpoints, no shell access, and no secret reveal controls.
- **Opt-in MCP interface for AI assistants:** Alongside the read-only web UI, an opt-in Model Context Protocol (MCP) interface lets AI assistants answer operational questions in a single prompt, run gated on-demand diagnostics (`ping`/`traceroute`), and perform audit-logged configuration changes exclusively through VyOS's native `set → commit-confirm → save` workflow.
- **Self-contained embedded footprint:** One static Go binary (`CGO_ENABLED=0`) plus one pre-allocated multi-ring data file (`vyview.ring`), powered by Datastar, UnoCSS, Font Awesome Pro Sharp Light (with Free fallback), IBM Plex Sans/Mono, and uPlot.
- **Hardware targets:** Reference platform is the **NXP LS1046A "Mono Gateway"** (4× Cortex-A72, `arm64`, Linux 6.18, VyOS 1.5/rolling, eMMC/SD flash), with full functional and CI parity on stock `x86_64` VyOS. Optional platform modules declare capabilities and degrade gracefully when absent.

---

## 2. Goals & Non-Goals

### Goals
1. **Best-in-class operational GUI:** Match or exceed commercial router dashboards (UniFi, FortiGate, Verso) in visual clarity and polish, while surpassing them in source transparency, plain-language explanations, and historical depth. Replace generic "Records" scaffolding with purpose-built operational views.
2. **Answer plain-language questions in ≤ 3 clicks (or 1 MCP prompt):** Support both technical diagnostics and judgement questions (*"Is my firewall safe?"*, *"Am I protected well enough?"*, *"Why is my internet not working?"*, *"What devices are visible to my router?"*).
3. **Hardware-agnostic acceleration & OT readiness:** Provide a unified **Dataplane / Acceleration** section (ASK2, VPP, Flowtables, AF_XDP, CAAM, QAT), a dedicated **Redundancy** section (VRRP, `conntrack-sync`, WAN failover, MRP rings), and IT/OT change traceability without harming real-time traffic latency.
4. **Lower the barrier for users migrating from other brands:** Present firewall policies in a familiar FortiGate/OPNsense policy-table layout, include multi-vendor CLI cheat sheets (Cisco, Juniper, Fortinet, MikroTik, Hirschmann, Siemens SCALANCE), and support MCP-assisted migration diffs.
5. **Offline LAN HTTPS by default:** Serve cleanly over HTTPS on the LAN or management VRF with zero CDN dependencies so browser features such as Clipboard API (`Copy`) work reliably everywhere.

### Non-Goals (Web GUI)
- **No configuration or runtime mutations in the web UI** (writes belong strictly to VyOS CLI/API or opt-in MCP `commit-confirm`).
- **No interactive shell or arbitrary command runner.**
- **No built-in outbound alert delivery** (email/Slack/PagerDuty). Instead, point Prometheus at vyview's recommended `/metrics` endpoint — which exports both telemetry and vyview's own health checks/alerts — and let Alertmanager handle delivery (optionally as a VyOS Podman container).
- **No multi-router fleet management** and **no internet-facing exposure**.
- **No heavy on-box engines** (no nDPI packet inspection, InfluxDB, ClickHouse, SQLite WAL, or Grafana on the router).

---

## 3. Five Binding UX Design Principles

Every view and widget across vyview follows five human-centered rules, backed by engineering constraints in [Design System](../ui/design-system.md) and [Performance & Testing](performance-and-testing.md):

1. **Every page answers one question** — Build purpose-built views tailored to the operator's question rather than repeating generic "Records" tables.
2. **Explain, don't just show** — State *why* an item is warning, *why* a flow is not hardware-offloaded (e.g., short-lived UDP below `hardware-offload-delay 500 ms`), and *what to do next*.
3. **One truth everywhere** — Never duplicate the same dataset across multiple panels on one page, and ensure metrics reconcile identically across linked pages (for example, the `Errors` KPI tile on Interface Overview matches `rx_fcs_err` in Counters; Interface Queues matches QoS CAKE tins; firewall default-policy drops sum to the Dropped Packets tile).
4. **From glance to depth** — Lead with color + icon shape + a one-line plain-language health summary, placing deep engineering telemetry and raw counters below or behind tabs and drawers.
5. **No magic numbers** — Always pair raw hex codes, register fields, OIDs, and thresholds with plain-language meaning (e.g., eMMC `0x03` → `"20–30% worn (70–80% life remaining)"`, `"Lease 21 h remaining"`).

*Supporting engineering principles:* read-only web GUI by construction; source-transparency chips on every widget; shared URL-backed page time range (`$range`) with synced crosshairs; live-by-default with pause + pending count; comfortable/compact/wall densities; WCAG 2.2 AA accessibility.

---

## 4. Target Personas

| Persona | Role & Primary Needs |
| --- | --- |
| **P1 — Owner / Architect** | Validates VyOS commits, kernel updates, and hardware-offload changes; inspects exact counters, diffs, and resource budgets. |
| **P2 — Troubleshooting Power User** | Follows live symptoms (latency spikes, drops, BGP flaps, VPN drops) from Overview to root-cause evidence in seconds. |
| **P3 — Non-Technical / Home & SMB Admin** | Glances at plain-language health summaries, WAN connectivity checks, and top bandwidth users without touching the CLI. |
| **P4 — Kiosk / Wall Display User** | Runs long-lived unattended rotation on a 1920px+ display with high legibility and zero animation fatigue. |
| **P5 — Developer / Contributor** | Builds and verifies changes on a laptop using deterministic fake collectors, golden parser fixtures, and automated Playwright/axe gates. |
| **P6 — OT Engineer (IT/OT Boundary)** | Manages plant/office segmentation and edge routers under strict change control where downtime is unacceptable. Needs IT/OT zone visibility, audited change traceability (who/when/via CLI-API-MCP), L2/TSN/IP industrial protocol visibility (EtherCAT, PROFINET, OPC UA, EtherNet/IP, Modbus TCP), latency/jitter determinism, and `commit-confirm` safety. |

---

## 5. Core Plain-Language Question Workflows

Every common operational question is answerable in **≤ 3 clicks in the GUI** and **1 prompt via MCP**:

| Question | GUI Path (Clicks) | Evidence Surfaced |
| --- | --- | --- |
| *"Is everything OK, and what changed?"* | Overview (`0`) | Top-bar Health pill + plain-language sentence, WAN card, 7 KPIs, dated recent events |
| *"Why is my internet not working?"* | Overview → WAN card / Diagnostics (`1`) | Step-by-step chain highlighting first failing step (`Link → DHCP/IP → Default route → Gateway RTT/loss → DNS → External probe`) + recent WAN events & last commit |
| *"Why is … not working?"* (device, site, VPN) | `Ctrl/Cmd-K` → query (`1`) → match (`2`) | Device: lease, ARP/NDP, port, matching firewall rules/drops. Site: DNS, conntrack, NAT, drops. VPN: handshake/SA state |
| *"What devices are visible to my router?"* | Traffic & Clients → Inventory (`1`) → device (`2`) | Per-segment/VLAN OT asset inventory: name, IP, MAC, OUI vendor, identity source (DHCP/ARP/NDP/LLDP), first/last seen, detected protocols, traffic |
| *"Is my firewall safe?"* | Firewall & NAT (`1`) → Security check (`2`) | Read-only pass/fail audit: WAN default drop, no exposed WAN services, port forwards + rules, zones without policy, never-matched rules, DNS-bypass detection |
| *"Am I protected well enough?"* | Health pill (`1`) → Security (`2`) | Combined posture score: firewall check + default password / factory SSH key detection, SSH management-only, image age, cert expiry, unexpected open ports |
| *"If something fails, what takes over?"* | Redundancy (`1`) | VRRP master/backup, `conntrack-sync` state, WAN failover path, MRP ring status |

*(Additional operator jobs from the original brief — slow internet/bufferbloat, bandwidth consumers, VPN state, commit impact, hardware offload verification, IoT→NAS reachability, BGP health, hardware/thermal status, and flash wear — follow the same ≤ 3-click rule.)*

---

## 6. Success Metrics

| Metric | Target & Verification Method |
| --- | --- |
| **Time to answer (GUI)** | ≤ 3 clicks or one search; verified by Playwright journeys and 3-user usability tests |
| **Time to answer (MCP)** | Every core job answered in 1 question; scripted MCP suite verified against GUI evidence |
| **MCP write safety** | `100%` of writes via `commit-confirm` and `100%` audit-logged; automated write-safety test suite |
| **First paint (LAN)** | Warm `< 300 ms`, cold `< 600 ms`; measured via Playwright tracing |
| **Idle CPU (LS1046A)** | `< 3%` of one Cortex-A72 core (1 Overview client, 10-minute `pidstat`) |
| **Real-time / TSN no-harm** | Running vyview adds `< X µs` (TBD by hardware measurement) to p99 forwarding latency and jitter under load |
| **Memory (RSS)** | `< 40 MB` (`GOMEMLIMIT=40MiB`) after 72-hour soak, slope `< 1 MB/day` |
| **Flash writes** | Batched rollups only; `< 50 MB/day` host writes measured over 7 days |
| **Update latency** | `< 1.5 s` end-to-end (`200 ms` live fast loop, `60 fps` `<vy-chart>` interpolation) |
| **MRP ring visibility** | `< 2 s` from ring open/closed to UI & timeline; `100%` of state changes captured vs. `bridge mrp` |
| **PTP / time sync** | Clock offset & sync state displayed at `1 s` cadence when present (verified vs. `pmc`/`phc2sys`) |
| **OT asset discovery** | `≥ 95%` of devices on directly connected segments discovered on test network |
| **Hardware parity** | All Playwright journeys pass on both `x86_64` (no offload) and `LS1046A` (with ASK2) |
| **Accessibility** | WCAG 2.2 AA (`0` serious or critical `axe-core` violations) |
