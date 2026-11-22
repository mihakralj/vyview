# Diagnostics

> **Primary Questions:** *"Why is my internet not working?"* and *"Is the path to the internet healthy over time?"* · **Route:** `/diagnostics`

---

## Layout & Key Views

Built around scheduled, fixed-target cached probes (keeping the web GUI strictly read-only; ad-hoc on-demand `ping`/`traceroute` run via the opt-in **MCP** interface):

- **Step-by-Step WAN Connectivity Chain (Top):** Visual green/amber/red diagnostic chain highlighting the **first failing step**:
  ```text
  Link Up → DHCP / IP Lease → Default Route → Gateway Reachable (RTT/Loss) → DNS Resolving → External Internet Probe
  ```
  Paired directly with recent WAN events and the latest configuration commit so outages correlate with changes immediately.
- **Explicit Probe Context & Historical Trends:** Every probe card states **exactly what it tests** (target IP/host, DNS query name and resolver), the **threshold** it is judged against, **consistent units** (`ms`, `%`, `Mb/s`), **full dated timestamps with timezone**, and **historical trend charts** (latency, jitter, and packet loss over time — never isolated single-point numbers).
- **Path Diffs (`mtr`) & Optional Speed Tests:** Displays `30 d` traceroute hop changes and optional scheduled nightly speed tests (off by default, `90 d` retention).

---

## Data Sources & Cadence

- **ICMP / UDP RTT & loss (`10 s`):** Unprivileged Go ICMP datagram sockets (`ping_group_range`), hard rate-limited (`≤ 1 probe/target per 10 s`, max `8` configured targets).
- **DNS resolution (`60 s`):** Fixed configured test query against local/upstream resolvers.
- **Traceroute (`15 min`):** Allowlisted `mtr --json` against fixed configured targets only (never user-supplied inputs).

---

## UX & Consistency Constraints

- Never accept arbitrary user-typed hostnames or flags in the web GUI.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 23–24).
