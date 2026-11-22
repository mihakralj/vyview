# Traffic & Clients

> **Primary Questions:** *"Who is using bandwidth?"* and *"What devices are visible to my router?"* (OT asset inventory) · **Routes:** `/clients`, `/clients/{id}`, `/conntrack`

---

## Layout & Key Views

Uses **one rich inventory table plus purpose-built visualizations** — never repeating client data 3× across separate "Client identity evidence", "At a glance", and "Records" blocks:

- **Top Talkers Chart:** True **per-client stacked area chart** (top 6 clients + `other`) with Y-axis values and time labels, synchronized strictly to the page time range.
- **Client & OT Asset Inventory Table:** Displays **both RX and TX** rates/totals (uploads, backups, and exfiltration matter as much as downloads), IP/hostname, MAC + OUI vendor, privacy-randomized MAC badge, **connection type** (Wired vs. Wi-Fi, kept separate from online/idle status and disclosing its source such as AP integration or LLDP), switch port / AP, identity source (`DHCP`, `ARP`, `NDP`, `LLDP`), **first seen & last seen**, trend sparkline, **detected protocols** (including OT protocols: Modbus TCP, PROFINET, OPC UA, EtherNet/IP), hardware-offload indicator, and top destinations. Always state explicit cardinality counts (e.g., `"Showing top 50 of 47 active clients"`).
- **Traffic Paths (Sankey):** Lazy-loaded ECharts Sankey showing `WAN / VPN ↔ per-segment/VLAN` (`LAN`, `IoT`, `Guest`, `Lab`) **and inter-VLAN / IT↔OT flows** directly correlated with firewall zones. Omit meaningless constant status columns like `"Observed"`.
- **Conntrack & NAT Table (`/conntrack`):** Dense virtualized table (refreshes every `5 s` **only while open**; no kill/flush controls) with validated quick filters (client, protocol, state, path, port), top-N by traffic, inline **NAT translation column** (`192.0.2.20 → 203.0.113.42`), **client/DNS name resolution** (`nas-01 → github.com`), packets/bytes/age/timeout, hardware offload shown as **`%`** (`"80% via hardware"`) with explicit path (`HW` / `SW-flowtable` / `VPP` / `kernel`), link to the **matching firewall rule**, both `early_drop` and `insert_failed` table-pressure counters, and a plain-language **"Why isn't this flow offloaded?"** explanation (e.g., short-lived UDP below `hardware-offload-delay 500 ms`). Flags direct external DNS flows bypassing the router's resolver for the Firewall Security Check.

---

## Data Sources & Cadence

- **Two-Plane Per-Client Accounting (`10 s`):** Aggregates `CTNETLINK` `NEW`/`DESTROY` events and periodic stats, reconciled with the ASK2 hardware flow table (`ynl --family ask`) by client IP so hardware-offloaded flows are never under-reported.
- **Bounded Cardinality:** Tracks top `50` clients + `other` via a constant-RAM **Space-Saving / Heavy-Hitter Top-K sketch**.
- **Identity Enrichment (`30–60 s`):** Kea DHCP leases, `RTM_GETNEIGH`, `lldpcli`, and OUI database.

---

## UX & Consistency Constraints

- Respect `privacy.per_client = off | aggregate | full` (no DNS query logging).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 17).
