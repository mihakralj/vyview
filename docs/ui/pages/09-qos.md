# QoS

> **Primary Question:** *"Is the shaper preventing bufferbloat under load, and how much traffic actually passes through it?"* · **Route:** `/qos`

---

## Layout & Key Views

Purpose-built shaping view with **no generic "Records" block or duplicate "At a glance" list**:

- **Both Ingress (Download) and Egress (Upload) Shapers:** Shows configured bandwidth, Bufferbloat grade (`A`–`F`), idle vs. loaded latency (`8.4 ms idle → 16.2 ms loaded`), and scheduled latency-under-load probe history (`90 d`).
- **All Configured CAKE Tins (3, 4, or 8 tins — including empty ones):** Renders every tin (e.g., `diffserv4`: `Bulk`, `Best Effort`, `Video`, `Voice`) so traffic classification is transparent. **Warn on elevated delay in Voice/Video, never on Bulk** (delay in Bulk is intentional by design).
- **Multi-Tin Delay & AQM Signals:** Multi-tin delay chart with explicit Y-axis and time labels, displaying **both ECN marks and packet drops** (`CAKE` commonly signals congestion via ECN rather than drops).
- **Hardware-Offload Bypass Transparency:** When ASK2 or `nf_flowtable` offload is active, established flows bypass the CAKE qdisc. Explicitly display the **`%` of interface traffic traversing CAKE vs. bypassed via hardware offload** so a `"Bufferbloat: A"` badge is never misleading.

---

## Data Sources & Cadence

- **Qdiscs & CAKE `xstats` (`5 s`):** Direct kernel netlink (`florianl/go-tc` `RTM_GETQDISC` / `RTM_GETTCLASS`, JSON `tc` fallback) because VyOS op-mode output names vary across releases.
- **Latency-under-load history (`90 d`):** Stored in `vyview.ring` Event/Probe ring.

---

## UX & Consistency Constraints

- Queue backlog, ECN mark, and drop totals must reconcile identically with the parent interface's **Queues** tab (**One Truth**).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 20).
