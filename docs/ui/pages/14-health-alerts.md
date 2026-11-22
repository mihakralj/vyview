# Health & Alerts

> **Primary Question:** *"What is wrong right now, since when, and am I protected well enough?"* · **Route:** `/health`

---

## Layout & Key Views

- **Active (Engaged) vs. Past Alerts (`90 d`):** Plain-language status banner at top, list of currently active issues with start timestamp, current value vs. threshold, direct evidence link, and `"what to do next"` guidance, followed by searchable 90-day alert transition history.
- **Built-In Operational Checks:**
  - WAN down or gateway packet loss `> 5%` for `2 min`
  - **VRRP master/backup flap** & **MRP ring open/flap**
  - **Local control-socket / daemon failure** (`Kea`, `VICI`, `Podman`, `D-Bus`, `vyview-probe`)
  - BGP neighbor down `> 2 min`; expected WireGuard peer stale `> 3 min`
  - Disk usage `> 85%`, storage target unavailable/insufficient space, or **eMMC wear `> 50%` advisory (`≥ 0x06`) / high (`≥ 0x08`) / `Pre-EOL ≥ 0x02`**
  - High temperature (explicit sensor), SFP+ DOM Rx/Tx optical power out of range, certificate expiry `< 21 days`
  - DHCP pool `> 90%`, conntrack `> 75%` / `> 90%` (`early_drop` / `insert_failed`), BMan buffer pool `< 30%` / `< 15%`
  - Failed systemd unit; NTP unsynchronized `> 10 min` or PTP offset out of bounds
- **Read-Only Security Check Tab (*"Am I protected well enough?"*):** Combined security posture score aggregating the [Firewall Security Check](06-firewall-nat.md) (WAN default drop, port forwards, exposed services, DNS-bypass detection) plus **default password / factory SSH key detection**, SSH management-interface-only verification, VyOS image age, certificate expiry, and unexpected listening ports (`NETLINK_SOCK_DIAG`).

---

## Data Sources & Cadence

- Derived in-process from collectors (`5–60 s`) plus event-driven triggers (`rtnetlink`, `keepalived` D-Bus, `bridge mrp`, `journald`).
- **Prometheus `/metrics` export:** All built-in health checks and alert states are exported on vyview's recommended `/metrics` endpoint so Prometheus + Alertmanager (optionally as a VyOS Podman container) handle outbound notifications.

---

## UX & Consistency Constraints

- Alert acknowledgments are browser-local (`localStorage`); alert thresholds are configured outside the GUI and shown read-only in plain language.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 23).
