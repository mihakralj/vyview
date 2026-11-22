# Interface Detail

> **Primary Question:** *"How healthy is this specific port, where are drops occurring, and how is it configured?"* · **Route:** `/interfaces/{name}`

---

## Layout & Key Views

Purpose-built per-interface tabs with **no generic "Records" scaffolding**. Clicking any counter, optic reading, or queue opens a **right-hand detail drawer on desktop** (full-screen sheet on phones) with real entity URL IDs (`?sel=rx_fcs_err`), labelled fields, an embedded trend chart, hardware-specific badges (e.g., `DPAA/LS1046A`), descriptive evidence links (e.g., `"View BMan pools in Acceleration"`), plain-language explanations, and no Copy button on empty (`—`) values.

- **Overview Tab:**
  - Plain-language health summary banner (*"This port is healthy. 1 link interruption in the last 24 h, 54 corrupt packets received."*)
  - Identity, driver, MAC, speed/duplex, **LLDP neighbor card** (remote device + port), and MRP ring role/state on ring ports
  - KPI tiles reconciled with driver counters (**One Truth:** `Errors` tile must match `rx_fcs_err`/CRC counters; **Link flaps** clickable to the exact flap event)
  - True mirrored **RX-above / TX-below** throughput chart with Y-axis values and time labels, per-port hardware `offload_ratio`, optics DOM gauges with warning/alarm zones, and clickable driver counters with plain-language tooltips
- **Counters Tab:**
  - Complete counter table (never fewer counters than Overview), **explicitly splitting drops by source**: `netdev` vs. `ethtool`/DPAA1 FMan MAC & RX FIFO overflow vs. `/proc/net/softnet_stat` vs. BMan depletion
  - Trend sparkline column, period-labelled delta header (`Delta (1 h)`), and clickable rows matching Overview
- **Optics Tab *(shown only on SFP/SFP+ ports)*:**
  - Module identity (vendor, part number, serial, transceiver type such as `10GBASE-LR/SR`, wavelength, DOM support)
  - Visual gauges with **both warning and alarm limits**, plus historical **RX/TX power and temperature trends** to detect degrading optics or dirty fiber early
  - Omit unnecessary filter/sort toolbars on compact 5-row telemetry tables
- **Addresses Tab:**
  - Strictly separates the interface's own IPv4/IPv6 addresses (including **IPv6 link-local**, **prefix delegation**, and **DHCP lease details**: server, total lease, remaining expiry) from **Gateway / Neighbor reachability** (never list the gateway as the interface's own address; reserve `"Reachable"` for neighbors/gateways; omit Copy buttons on static labels like `"DHCP"`)
- **Queues Tab:**
  - Displays **both ingress and egress software shapers** (CAKE bandwidth, all tins, ECN marks vs. drops, delay, link to [QoS](09-qos.md)) **and NIC hardware queues** (DPAA frame queues / CEETM), clearly labelled SW vs. HW
  - Shows the `%` of port traffic traversing CAKE vs. bypassing via hardware offload, explaining that CAKE drops over a named window are normal AQM behavior
- **Config Tab:**
  - Full read-only VyOS config subtree as one copyable `set`-command block (`interfaces ethernet eth0 ...`, firewall zone, VRF, offload), **configured vs. running drift indicators**, and link to the **last commit diff** touching this interface

*Type variants:* Add bridge FDB/VLAN tables, bond/LACP partner state, PPPoE session details, WireGuard peers, tunnel endpoints/VNI, wireless stations, or loopback addresses as appropriate.

---

## Data Sources & Cadence

- **Throughput & link state:** `200 ms` `IFLA_STATS64` + event-driven `rtnetlink` flaps.
- **Driver/MAC counters & queues:** `1–5 s` `ETHTOOL_GSTATS`, `RTM_GETQDISC`, `softnet_stat`, and Two-Plane Reconciler.
- **Optics DOM EEPROM:** `300 s` via `vyview-probe` (slow I2C bus).
- **LLDP & config:** `60 s` `lldpcli` / on-commit config archive.

---

## UX & Consistency Constraints

- Hide the **Optics** tab on copper RJ45 and virtual interfaces.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 16).
