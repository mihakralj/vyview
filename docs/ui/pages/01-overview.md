# Overview

> **Primary Question:** *"Is everything OK, and what changed?"* · **Route:** `/` (kiosk: `/?kiosk=1`)

---

## Layout & Key Views

Designed to fit within a single `1440 × 900` desktop viewport with **no generic "Records" scaffolding**:

- **Top-Bar Health Pill & Plain-Language Summary:** Active warnings and alerts surface in the top-bar Health pill alongside a single plain-language health sentence linking directly to [Health & Alerts](14-health-alerts.md).
- **WAN Status Card:** Displays link state, IPv4/IPv6 addresses, gateway, RTT/loss, explicit lease wording (`"Lease 21 h remaining"` vs. total duration), failover state, and 1-click entry to the step-by-step *"Why is my internet not working?"* chain in [Diagnostics](15-diagnostics.md). Secondary standby or down WAN links (e.g., `pppoe0`) must show an explicit `"Standby"` or `"Down"` colored dot + label rather than `"–"` with a neutral dot.
- **Hero Throughput Chart:** Live `200 ms` (5 Hz) netlink rates rendered at `60 fps` in `<vy-chart>` with tier history, event annotations, anomaly band, synchronized cursor, and export. Label directions in human **Download / Upload** terms (not ambiguous per-interface `↓/↑` where LAN `↑` means download) and explicitly note when LAN throughput exceeds WAN throughput due to local NAS or inter-VLAN traffic.
- **Seven Unambiguous KPI Cards:** CPU (`> 80%` for 5 min warns), memory, temperature (explicit sensor name: CPU / PCB / SFP), conntrack (`12.4k of 262k`; warns `> 75%`, crit `> 90%`), clients (report `"Wi-Fi"` only when backed by local Wi-Fi or AP integration, disclosing the source), VPN, and optional Acceleration/Offload (explicitly stating whether `%` counts flows, packets, or bytes).
- **Linked Top Talkers & Recent Events:** Top 5 clients + `other` (each linking directly to client detail) and the 10 most recent events with full date + time + timezone timestamps linking to their affected objects.

---

## Data Sources & Cadence

- **Throughput & CPU:** `200 ms` fast loop (`RTM_GETLINK` `IFLA_STATS64`, `/proc/stat`).
- **WAN state & RTT:** Link/failover state `5 s` (plus event-driven `rtnetlink` flaps), gateway RTT/loss `10 s`.
- **KPIs & events:** `5–10 s` snapshot rollups and event ring.

---

## UX & Consistency Constraints

- Never place `"On this page → Links"` anchor items inside the left application rail.
- Every KPI figure must reconcile identically with its destination page (**One Truth**).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 14–15).
