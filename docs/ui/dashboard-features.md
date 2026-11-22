# Dashboard Features, Cross-Page Widgets & Mockups

> Covers shared interactive behaviors, `60 fps` live charting, time-range tier mapping, reusable cross-page widgets, and the visual mockup index (`M1`–`M16`).

---

## 1. Interactive Behaviors

The server owns all data and HTML rendering; the browser holds only ephemeral interaction state:

- **Global Search & Command Palette (`Ctrl/Cmd-K`):** Cross-entity search across interfaces, clients, routes, rules, VPN peers, config paths, and VyOS CLI commands (`show interfaces`), with keyboard navigation and `< 50 ms` p95 response target.
- **Shared Time Range & Comparison:** Preset windows (`5 m`, `1 h`, `24 h`, `7 d`, `30 d`, `90 d`) plus custom absolute `[from, to]`, mirrored in the URL (`?range=24h`). Supports previous-period comparison overlays (`?cmp=1`), shared event annotation markers, synchronized crosshairs (`cursor.sync.key = page`), and brush-to-zoom.
- **Live & Pause:** Live streaming is enabled by default. Pressing `Space` or clicking Pause sets `$live=false`; the server patches only `#pending` counts while paused and sends one clean snapshot upon resume. A local `1 s` client ticker flags widgets stale when `age > 2 × cadence`.
- **`60 fps` Live Charts (`<vy-chart>` + uPlot):** Decouples smooth `60 fps` `requestAnimationFrame` X-axis scrolling (`250–300 ms` jitter buffer + leading-edge interpolation) from the `200 ms` (5 Hz) telemetry tick. Rolled-up tiers render the sustained `avg` line with a translucent `max` peak envelope and optional hour-of-week anomaly bands.
- **Exports, Deep Links & Kiosk:** Copy-as-VyOS-CLI buttons, CSV/JSON/PNG widget export, URL-shareable drawer selections (`?sel=`), saved views (JSON import/export), keyboard shortcut sequences, kiosk rotation mode, and installable offline PWA shell.

### Time-Range to Storage-Tier Mapping

| Requested Window | Storage Tier | Resolution | Notes |
| --- | --- | --- | --- |
| `≤ 6 min` (`2 m` / `5 m` live) | **T0** (RAM) | `200 ms` (5 Hz) | Streamed live, `60 fps` interpolated in `<vy-chart>` |
| `≤ 24 h` (`15 m`, `1 h`, `6 h`, `24 h`) | **T1** (`tmpfs`) | `5 s` | Seamless fallback to `T2` across reboot gaps or in compact profile beyond `2 h 51 m` |
| `≤ 14 d` (`7 d`, `14 d`) | **T2** (`vyview.ring`) | `60 s` | `300 s` up to `7 d` in compact profile |
| `≤ 90 d` (`30 d`, `90 d`) | **T3** (`vyview.ring`) | `300 s` | Warmed into bounded `≤ 4 MiB` RAM query cache on read |
| `> 90 d` (up to `2 y`) | **T4** (`vyview.ring`) | `3600 s` | Feeds hour-of-week anomaly bands (`median ± k·MAD`) |

*Point budget:* The server caps every chart query to `~1,000–1,500` points per series via min/max-preserving M4/LTTB downsampling.

---

## 2. Reusable Cross-Page Widgets

| Widget | Standard Behavior & Accessibility |
| --- | --- |
| **Source Chip** | Present on every widget; shows the copyable VyOS CLI equivalent, collection mechanism badge (`netlink`, `/proc`, `vtysh`, socket), cadence, last success time, and parser version. |
| **KPI Card** | Displays headline value (`tabular-nums`), unambiguous unit/label, period delta, inline SVG sparkline, and 5-state status badge (color + icon + text). |
| **Time-Series Chart (`<vy-chart>`)** | Synchronized cursor across page charts, event annotations, dashed threshold lines, anomaly band toggle, and an accessible HTML data-table alternative. |
| **Virtualized Table** | Sticky header, fixed row height (`+10` overscan), server-side filtering and sorting above `2,000` rows, persisted column chooser. |
| **Right-Hand Detail Drawer** | Opens on the right on desktop (full-screen sheet on mobile) with independent shareable URL (`?sel=<entity_id>`), trapped focus, `Escape` to close, labelled fields, embedded trend chart, and no Copy button on empty (`—`) or `[REDACTED]` values. |

---

## Mockups

All mockup data and ASK2 values are illustrative; public IPs use RFC 5737 / RFC 3849 documentation ranges. Each link opens a PNG rendering of the corresponding page from the [original design brief](../reference/VyOS-ViewOnly-Dashboard-Design-Brief.pdf) (pp. 31–46; callout text in [brief-extracted.txt](../reference/brief-extracted.txt)).

| Mockup | Key Visual Callouts |
| --- | --- |
| [M1 — Overview](../reference/mockups/m01-overview.png) | Conditional alerts; WAN evidence; event markers and anomaly band; KPI status; dated events |
| [M2 — Interfaces](../reference/mockups/m02-interfaces.png) | Multi-select chips; columns; one-hour sparklines; error highlights; source footer |
| [M3 — Interface Detail](../reference/mockups/m03-interface-detail.png) | Tabs; mirrored RX/TX throughput and flap; errors; DOM thresholds; driver counters |
| [M4 — Traffic & Clients](../reference/mockups/m04-traffic.png) | Top 6 + other stacked chart; Sankey; identity sources; randomized MAC privacy badge |
| [M5 — Firewall](../reference/mockups/m05-firewall.png) | Chains; rule sparklines and shorthand; zone matrix; rule detail drawer |
| [M6 — Conntrack & NAT](../reference/mockups/m06-conntrack-nat.png) | Offload totals; validated filters; virtual rows; HW/SW path; flow detail drawer |
| [M7 — Routing](../reference/mockups/m07-routing.png) | Neighbor/BFD state; diagnosis; longest-prefix match; RIB churn |
| [M8 — WireGuard / VPN](../reference/mockups/m08-wireguard.png) | Optional acceleration; stale thresholds; public keys only; handshake age chart |
| [M9 — QoS](../reference/mockups/m09-qos.png) | Bufferbloat grade; CAKE tins; latency-under-load comparison; backlog |
| [M10 — Dataplane / ASK2](../reference/mockups/m10-ask2.png) | Capability notice; HW/SW flow split; CEETM heatmap; BMan pool; Suricata shunt |
| [M11 — System](../reference/mockups/m11-system.png) | CPU heatmap; translated eMMC wear; installed images and boot history |
| [M12 — Configuration](../reference/mockups/m12-configuration.png) | Collapsible tree; diff modes; masked secret changes; offline explain |
| [M13 — Logs & Events](../reference/mockups/m13-logs.png) | Severity histogram; paused tail; pending count; cursor permalinks |
| [M14 — Command Palette](../reference/mockups/m14-palette.png) | Cross-entity query; client hit; state/config evidence; keyboard hints |
| [M15 — Mobile Layout](../reference/mockups/m15-mobile.png) | Alert banner; three-column KPIs; bottom navigation; range picker; copyable detail |
| [M16 — Topology](../reference/mockups/m16-topology.png) | Router/offload state; utilization & speed bands; dashed tunnels; LLDP/BGP health |
