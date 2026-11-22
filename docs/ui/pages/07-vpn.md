# VPN

> **Primary Question:** *"Are my VPN tunnels and peers up, encrypted in hardware where supported, and passing traffic?"* · **Route:** `/vpn`

---

## Layout & Key Views

Unified capability-detected VPN hub (hiding unconfigured tunnel technologies):

- **WireGuard:** Interface and peer tables with endpoint, allowed IPs, last handshake age (with age trend chart), RX/TX rates reconciled with ASK2 fast-path RX (`wgfp_rx`), and optional hardware-acceleration indicator. Expected-traffic peer stale thresholds: **warn `> 3 min`**, **critical `> 15 min`**.
- **IPsec (strongSwan / XFRM):** Active IKE & child SAs, tunnel state, rekey timers, replay/authentication error counters (`/proc/net/xfrm_stat`), and **CAAM / QAT hardware crypto verification** showing whether encryption runs in silicon or falls back to software.
- **OpenVPN:** Active server/client instances, connected clients, and throughput.
- **Optional Overlay VPNs:** Tailscale / ZeroTier status cards shown only when installed and configured.

---

## Data Sources & Cadence

- **WireGuard (`5 s`):** `wgctrl` generic netlink (`vyview-probe`) reconciled with ASK2 `wgfp_rx`.
- **IPsec (`10 s`):** Read-only strongSwan VICI socket (`/run/charon.vici`, `list-sas` only) + `NETLINK_XFRM` + `/proc/net/xfrm_stat` + CAAM/QAT crypto job counters.
- **OpenVPN (`30 s`) / Tailscale & ZeroTier (`60 s`):** Allowlisted read-only status queries.

---

## UX & Consistency Constraints

- **Public keys only:** WireGuard private keys and preshared keys are dropped at the collector boundary and never stored, indexed, or rendered. IPsec PSKs and private keys are never read.
- WireGuard interface lists must agree identically with [Interfaces List](02-interfaces-list.md) (**One Truth**).
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 19).
