# Interfaces List

> **Primary Question:** *"Are links up, at what speed, and how are they structured?"* · **Route:** `/interfaces`

---

## Layout & Key Views

Structured as a clean three-layer view from physical hardware down to logical topology:

1. **Physical Port Panel (Top):** Visual front-panel port strip showing live status per physical port (`Up`, `Down`, `Standby`), negotiated speed, RJ45 vs. SFP+ cage position, and WAN badge — matching known hardware layouts (e.g., Mono Gateway) or auto-generating a clean physical port strip on unknown hardware.
2. **Core Inventory Table:** Multi-select interface-type filter chips with counts, sticky headers, sortable columns, column chooser (persisted in `localStorage`), and 1-click IP address copy. Each row displays:
   - Status with **colored dot/icon + text** (`Up`, `Standby`, `Admin down`, `No carrier`)
   - Name, type, description, and **master / bridge members** by default
   - **Both IPv4 and IPv6 addresses**
   - MTU and negotiated speed/duplex
   - Live RX/TX rates, clearly labelled 1-hour sparkline (`RX`, `TX`, or mirrored both), and period-labelled error/drop deltas (`"+54 (1 h)"`) linking straight to the interface's **Counters** tab
3. **Logical Hierarchy Tree:** Collapsible tree showing bridge membership (`br-lan → eth2, eth3`), VLAN sub-interfaces under their parent (`eth4 → eth4.3900`), and **MRP ring membership/role** when configured.

---

## Data Sources & Cadence

- **Rates:** `200 ms` netlink `IFLA_STATS64` (monotonic $\Delta t$).
- **Link state, addresses, speed & metadata:** Event-driven `rtnetlink` multicast (`Link/AddrSubscribe`) + `5 s` reconciliation.
- **Logical hierarchy & MRP:** `rtnetlink` master/linkinfo + `bridge mrp` netlink events.

---

## UX & Consistency Constraints

- Clicking any port in the physical strip, table row, or hierarchy tree opens [Interface Detail](03-interface-detail.md).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 15).
