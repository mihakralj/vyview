# Logs & Events

> **Primary Question:** *"What errors or state changes happened recently, and which router object was affected?"* · **Routes:** `/logs`, `/events`

---

## Layout & Key Views

Live `journald` stream and curated 30-day event timeline with **no duplicate Event timeline across tabs**:

- **Severity-Stacked Histogram:** Top bar chart with explicit Y-axis counts and dated X-axis labels stacked by severity (`crit`/`err`, `warn`, `notice`/`info`) so `"anything bad?"` is answered at a glance.
- **Structured Filtering:** One-click **severity** (`CRITICAL`, `ERR`, `WARN`, `INFO`) and **systemd-unit** filter chips alongside free-text search, pause with pending counter, copy line, and shareable journal cursor permalinks.
- **Clear Columns & Full Timestamps:** Separate **Severity** (`Critical`, `Warning`, `Info`) from **State** (`Active`, `Recovered`). Always show **full date + time + timezone** (honoring the user's Settings timezone/24 h preference) so timestamps stay unambiguous across 7 d / 30 d windows.
- **No Magic Numbers + Actionable Deep Links:** Replace vague messages like `"Handshake older than threshold"` or `"91% allocated"` with exact context (`"Handshake 18 m old (warn > 3 m)"`, `"LAN DHCP pool 91% used (182/200, 18 free)"`), link every event straight to its target object (`DHCP pool → Services`, `peer → VPN`), and include a brief `"what to do next"` hint.

---

## Data Sources & Cadence

- **Journal tail:** Single long-lived allowlisted `/usr/bin/journalctl -o json -f` child process (event-driven, keeps `CGO_ENABLED=0`).
- **Curated events (`30 d`):** Stored in `vyview.ring` Region 6 (commits, reboots, link flaps, VRRP/MRP failovers, BGP/OSPF/BFD transitions, DHCP/WG/systemd alerts).

---

## UX & Consistency Constraints

- Live log appends never animate or steal scroll focus when paused or scrolled up.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 22–23).
