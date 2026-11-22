# Firewall & NAT

> **Primary Questions:** *"What traffic is allowed or blocked, which rules are actually hit, and is my firewall safe?"* · **Routes:** `/firewall`, `/nat`

---

## Layout & Key Views

Presents VyOS `nftables` policies in a familiar **FortiGate / OPNsense policy-table layout** to lower the barrier for users coming from other vendors — with **no duplicate matrix tab or generic Records block**:

- **Zone Policy Matrix (Above or Beside Rule Table):** Clicking any directional zone cell (e.g., `LAN → IoT`) immediately filters the rule table below without page jumps (`?zone=LAN--IoT`), highlights the active cell, and displays **real drop/accept counters on the default policy rule** so default drops reconcile cleanly with the `"Dropped packets"` KPI tile.
- **One Dense Grouped Rule Table:**
  - Preserves strict numeric rule evaluation order by default.
  - Supports a **comfortable / compact density toggle** (40+ rules visible on wide screens), grouped by chain or zone pair with collapsible headers (`"WAN → LAN · 12 rules · default drop"`).
  - Per row: relative **hit bar** alongside packet/byte totals, **last hit** (`"2 min ago"`, `"never"`), **never-matched cleanup marker** (`"never hit in 30 d"`), and compact icons for `log` and `state` instead of wide text columns.
- **Read-Only Firewall Security Check Tab (*"Is my firewall safe?"*):** Fixed read-only pass/fail audit checklist with direct evidence links:
  1. WAN default-drop policy verified
  2. No unintended router services exposed to WAN
  3. Port forwards paired with their corresponding filter rules
  4. Zones without an attached policy
  5. Rules never matched in `30 d`
  6. Clients bypassing the router's DNS resolver
- **Groups, NAT/NAT66 & Flowtables:** Address/network/port/domain groups with where-used links; SNAT/DNAT/NAT66 rule tables; hardware/software flowtable offload counters (`[OFFLOAD]` / `[HW_OFFLOAD]`); links to the virtualized Conntrack & NAT view.

---

## Data Sources & Cadence

- **Rule & default-policy counters:** `google/nftables` netlink (`vyview-probe`) joined to VyOS rules by comment/handle every `5 s`. History capped at top `120` rules + pinned rules.
- **Flowtables:** `10 s`.

---

## UX & Consistency Constraints

- Default-policy drop totals and rule counters must reconcile identically between the Zone Matrix, Rule Table, and KPI tiles (**One Truth**).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 18).
