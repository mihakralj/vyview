# Routing

> **Primary Question:** *"How does the router reach destination X, and are dynamic routing neighbors healthy?"* · **Route:** `/routing`

---

## Layout & Key Views

- **VRF Selector & Per-VRF RIB/FIB Tabs:** Displays the `default` VRF first and exposes every configured named VRF as an explicit selectable tab/pill so multi-VRF IPv4/IPv6 routing tables remain unambiguous.
- **Longest-Prefix Match Lookup:** Interactive read-only IP/CIDR lookup showing the winning FIB route, next-hop interface, ECMP paths, protocol source, and plain-language forwarding explanation.
- **Protocol Sub-Views:**
  - **Summary & Churn:** Route counts by protocol, RIB/FIB sync health, and recent route churn timeline.
  - **Static Routes:** Configured vs. active status.
  - **BGP:** Neighbor state, uptime, prefix counts, received/advertised route counts (noting when soft-reconfiguration is required for received routes), and diagnosis notes for idle/active peers.
  - **OSPF / OSPFv3:** Areas, neighbors, DR/BDR roles, and on-demand LSDB detail.
  - **BFD:** Session states, timers, and fast flap history.
  - **Policy-Based Routing (PBR):** Routing tables and matching rules.

---

## Data Sources & Cadence

- **FRR (`vtysh` batched JSON):** `30 s` for BGP, OSPF, and route summaries; `5 s` for BFD.
- **Kernel FIB:** `RTM_GETROUTE` netlink (`5–30 s`).
- **Expensive tables (full RIB / BGP received prefixes / OSPF LSDB):** Fetched on demand only while viewed.

---

## UX & Consistency Constraints

- Never expose neighbor reset or route-clear buttons.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 17–18).
