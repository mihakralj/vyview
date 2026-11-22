# Interfaces List

> **Primary Question:** *"Are physical ports up at expected speeds, how do VLANs/bridges/PPPoE/VPNs layer on top of them, and what IPv4/IPv6 addresses are bound?"* · **Route:** `/interfaces`

---

## Layout & Key Views

Structured as a unified **two-layer view** with **no duplicate flat table + separate tree**:

1. **Physical Front-Panel Port Strip (Top):** Visual hardware port strip showing each physical port (`RJ45` vs. `SFP+` cage), live status dot + label (`Up`, `Down / No carrier`, `Standby`), negotiated speed (`1G`, `2.5G`, `10G`), description, and child sub-interface count badge — matching known hardware layouts (e.g., Mono Gateway: `eth0`–`eth2` 1G RJ45, `eth3`–`eth4` 10G SFP+) or auto-generating a physical strip on unknown hardware. Clicking a port opens [Interface Detail](03-interface-detail.md).
2. **Unified Collapsible Hierarchy Tree-Table (`Hierarchy` [default] | `Flat` toggle):**
   - **Group 1 — Physical & L2 Hierarchy:** Strict kernel L2 nesting derived deterministically from `IFLA_LINK` and `IFLA_MASTER`:
     - Physical port → VLAN sub-interfaces → PPPoE (e.g., `eth3` → `eth3.10` → `pppoe10`; `eth4` → `eth4.20`)
     - Bridge / Bond master → physical member ports (`L2 member`) + VLAN SVIs (e.g., `br-lan` → `eth2`, `eth3`, `br-lan.20`)
     - Disconnected physical ports (`eth1`, `eth2`) show explicit `Down (u/D)` / `No carrier` status and `—` for addresses
     - **MRP ring membership/role** badged inline when configured
   - **Group 2 — Tunnels & VPN Overlays (L3-routed):** WireGuard (`wg*`), IPsec (`xfrm*`/`vti*`), OpenVPN (`vtun*`), and VXLAN/GRE sit as an L3 overlay group (never falsely parented to a single NIC that may failover) with a clickable **live underlay egress chip** (`via pppoe10`, `via eth3`) resolved from the active FIB / peer endpoints / `source-interface`.
   - **Group 3 — Loopback & Control:** `lo`, `dum*`, and VRF control interfaces.
   - **First-Class Multi-Address Stack (MAC + IPv4 + IPv6):** Every interface displays its L2 hardware `MAC` address (when applicable) first in the `Addresses` stack — never as a separate wide column — followed in canonical order by `IPv4`, `IPv6 GUA`, `IPv6 ULA`, and `IPv6 LLA` with borderless aligned scope tags and inline copy icons. Point-to-point interfaces (`pppoe*`) show their remote `peer` inline (`10.99.50.16 ↔ 10.99.50.1/32`), leased addresses show remaining validity (`DHCP · 22h`), multiple addresses per scope stack cleanly, and `tentative` LLAs on down ports are omitted from the overview (shown in detail).
   - **Columns & Controls:** Multi-select type filter chips with counts, search (including type, MAC & MTU lookup), `Hierarchy | Flat` toggle, `"Show kernel internal"` toggle (hides `pim6reg`, `gre0`, `sit0` by default; treats kernel `state UNKNOWN` + `LOWER_UP` as Up), column chooser (persisted in `localStorage`), and **5 balanced columns**: **`Interface`** (green/red Sharp Light type icon + `<name> - <Type> (MTU <size>), <Up|Down>` + description), **`Addresses`**, **`VRF`**, **`Link`**, and **`Throughput (1h)`** (`btop`-style full-row-height mirrored bar chart from a shared centerline — RX rising up in `--chart-1`, TX falling down in `--chart-2` — with `RX` top-right, `TX` bottom-right, and directional error/drop warnings such as `2 rx err` / `12 rx drop` shown **only when nonzero** in the middle linking directly to Counters).

---

## Data Sources & Cadence

- **Rates:** `200 ms` netlink `IFLA_STATS64` (monotonic $\Delta t$). Physical parent rows show actual wire totals (encompassing child VLAN/PPPoE/tunnel bytes); child rows show logical sub-interface share without double-counting.
- **Link state, addresses, MAC, VRF, hierarchy:** Event-driven `rtnetlink` multicast (`RTM_GETLINK` `IFLA_LINK`/`IFLA_MASTER`, `RTM_GETADDR`) + `5 s` reconciliation.
- **VPN underlay resolution & MRP:** Active FIB route lookup (`RTM_GETROUTE`) + `wgctrl`/VICI + `bridge mrp` netlink events.

---

## UX & Consistency Constraints

- Clicking any port in the physical strip, table row, or underlay chip navigates directly to [Interface Detail](03-interface-detail.md).
- Keep `Hierarchy` as default and switch cleanly to `Flat` when sorting globally across all interface types.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 15).
