# Topology

> **Primary Question:** *"How are WAN, VPNs, VLANs, switches, APs, firewall zones, and client segments physically and logically connected?"* · **Route:** `/topology`

---

## Layout & Key Views

Auto-drawn **layered L2/L3 network map** (replacing any flat box tree or duplicate "Records" table):

- **Top-to-Bottom Layers:**
  1. `Internet / ISP`
  2. `WAN link` — displays any **BGP session as a badge/label on the WAN link** rather than a fake separate physical cable
  3. `Router` — summary badge (e.g., `"24 devices"`), offload state, and **dashed VPN tunnels** to remote peers/sites
  4. `Bridges & VLANs`
  5. `LLDP neighbor switches & APs` — showing remote device name and port
  6. `Collapsible per-segment client groups` — e.g., `"IoT · 9 devices"`, `"+18 more devices"` (linking to Client Inventory) with **device-type icons** (laptop, NAS, switch, AP, PLC)
- **Firewall Zone Shading & MRP Rings:** Shaded colored zone regions around segments (`WAN`, `LAN`, `IoT`, `Guest`, `OT/Plant`) tying directly to the Firewall Zone Matrix and making IT/OT segmentation visible at a glance. Renders **MRP rings as an actual ring** with the blocked link clearly marked once VyOS supports MRP.
- **Visual Signals, Speed Mode & Legend:** Clear legend (`solid = wired`, `dashed = Wi-Fi/tunnel`, thickness/color by utilization `< 50%`, `50–80%`, `> 80%`, red when down) plus a **negotiated link-speed color mode** (e.g., green `≥ 1 Gbit`, orange `< 1 Gbit` to spot bad cables negotiating `100 Mbit` immediately). Warnings explain themselves inline (e.g., naming the specific WireGuard peer with a stale handshake).
- **Interactions:** Pan/zoom, keyboard focus navigation, locally pinned node positions (`localStorage`), and a right-hand link/node detail drawer.

---

## Data Sources & Cadence

- **Graph structure (`30 s`):** Netlink links/bridges/VLANs/routes, LLDP (`lldpcli`), firewall zones, VPN tunnels, and `bridge mrp`.
- **Live link utilization & health (`5 s`):** Patched onto SVG link `stroke` attributes without full graph redraw.
- **Rendering stack:** Zero-JS server-rendered Sugiyama-layered SVG by default; modular lazy-loaded ECharts for complex/high-node topologies.

---

## UX & Consistency Constraints

- Never repeat topology links as a duplicate "Records" table below the canvas.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 24).
