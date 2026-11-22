# Services

> **Primary Question:** *"Are local router services, containers, and control sockets healthy?"* · **Route:** `/services`

---

## Layout & Key Views

- **Service & Control-Socket Health Strip (Top):** Displays live operational status for each enabled service alongside **explicit per-daemon control-socket health** (`Kea`, `VICI`, `Podman`, `D-Bus`, `vyview-probe`) so localized daemon failures surface immediately as clear degraded badges rather than empty tables.
- **Containers (Podman):** Read-only cockpit-podman-inspired inventory of VyOS `set container` workloads showing container status, image/tag, uptime, and CPU/memory/network usage (strictly zero start/stop/exec controls).
- **DHCP (Kea IPv4 & IPv6):** Pool utilization bars (warn `> 90%`) with exact used/free counts, active leases, and static reservations.
- **DNS (PowerDNS Recursor):** Query rate, cache hit ratio, SERVFAIL/NXDOMAIN rates, and upstream forwarders (zero per-client DNS query logging).
- **Time Sync (Chrony NTP & PTP):** Synchronized stratum, offset, jitter, sources (warn unsynchronized `> 10 min`), plus **PTP clock offset and sync state** when `ptp4l`/`phc2sys` is present.
- **Additional Router Services (shown when configured):** Active SSH sessions (read-only; no session kill), LLDP neighbor table, mDNS repeater, IPv6 Router Advertisements, Dynamic DNS (warn mismatch `> 15 min`), UPnP/NAT-PMP mappings, and SNMP summary (communities/secrets redacted).

*(High-availability failover services — VRRP, `conntrack-sync`, WAN failover, and MRP rings — live in the dedicated [Redundancy](18-redundancy.md) section.)*

---

## Data Sources & Cadence

- **Podman (`10–30 s`):** Read-only HTTP-over-UNIX `GET` against `/run/podman/podman.sock` or `/run/user/<uid>/podman/podman.sock`.
- **Kea (`30 s`):** JSON UNIX sockets (`/run/kea/dhcp4-ctrl.sock`, `dhcp6-ctrl.sock`).
- **PowerDNS (`30 s`), Chrony (`60 s`), LLDP (`60 s`), PTP (`1 s` when present):** Allowlisted read-only sockets/commands (`rec_control`, `chronyc`, `lldpcli`, `pmc`/`phc2sys`).
- **SSH sessions (`30 s`):** `/run/utmp` / `sshd` journal events (no `w` subprocess).

---

## UX & Consistency Constraints

- Never expose container lifecycle, lease-revocation, or SSH session-kill actions.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 19–20).
