# Design Brief v1.1 Engineering Analysis

> Structured technical analysis of `VyOS-ViewOnly-Dashboard-Design-Brief-v1.1-FA-Pro.pdf` (v1.1, 17 Nov 2026, 82 pages, prepared for Miha Kralj).

---

## 1. Scope & Product Boundaries

**Target platform:** Single VyOS router on LAN or management VRF — reference hardware **NXP LS1046A "Mono Gateway"** (4× Cortex-A72, `arm64`, Linux 6.18, VyOS 1.5/rolling, eMMC/SD flash) plus stock `x86_64` VyOS.

**Information architecture:** **15 top-level sections** (~70 deep-linkable views), grouped into:
- **Top:** Overview (home + kiosk variant), Health & Alerts
- **NETWORK:** Interfaces (list + per-type detail: ethernet, bridge, VLAN, bond, PPPoE, WireGuard, tunnels, wireless, loopback), Traffic & Clients, Routing (FRR RIB/FIB/BGP/OSPF/BFD/PBR/VRFs), Firewall & NAT (nftables chains, zone matrix, groups, NAT/NAT66, conntrack, flowtables), VPN (WireGuard, IPsec, OpenVPN, optional Tailscale/ZeroTier), QoS (CAKE/HTB/FQ-CoDel, bufferbloat probe), Topology (L2/L3 auto-map)
- **PLATFORM:** Services (Kea DHCP/v6, PowerDNS Recursor, chrony, SSH, LLDP, mDNS, RA, DDNS, UPnP, SNMP, WAN LB), System (CPU heatmap, PSI/load, memory, hwmon/thermal, storage & eMMC wear, processes, systemd units, kernel/modules, images/boots, NIC driver/firmware, IRQ/softirq, vyview self-health), Dataplane / ASK2 *(optional pluggable module — hidden when absent)*
- **INSPECT:** Configuration (redacted tree, set-commands, JSON, commit history, Myers diff, offline explain panel), Logs & Events (journal live tail, firewall log, auth log, curated event timeline), Diagnostics *(cached results of scheduled probes only)*
- **Footer:** Settings (100% browser `localStorage`; server holds zero per-user UI state)

**Explicitly excluded (by construction, not convention):**
- Configuration changes (`/configure`, `/config-file`, `set`/`delete`/`commit`/`save`)
- Operational mutations (`reset`, `clear`, `restart`, `reboot`, `poweroff`, conntrack flush, DHCP release)
- Secret/image management (`generate`, `add/delete system image`)
- Arbitrary command execution / shell / user-triggered `ping`/`traceroute` in v1
- Outbound alert delivery (mail/push/webhooks) and multi-router fleet management in v1

---

## 2. Eight Binding, Testable Design Principles

Every principle in Brief §3 carries an automated CI gate:

| # | Principle | Automated Verification Rule |
| --- | --- | --- |
| 1 | **Read-only by construction** | Route-table test fails if any data-plane route accepts anything other than `GET`/`HEAD`; allowlist test + static check ban `os/exec` outside the allowlisted runner; netlink helper allowlists only `*_GET*` message types. |
| 2 | **Source transparency** | Every widget displays a copyable source chip (CLI command + mechanism badge, cadence, last refresh, parser version); template lint fails any widget lacking a declared source. |
| 3 | **Overview → zoom/filter → details** | Every list row links to a deep-linkable detail view/drawer (`?sel=`), preserving breadcrumbs, time range, and filters. |
| 4 | **Progressive disclosure** | Default view shows 5–8 key fields; full counters/LSDB/raw JSON behind tabs, toggles, and column choosers (`localStorage`). |
| 5 | **One shared time range** | Single `$range` Datastar signal mirrored in URL (`?range=24h` or `?from=&to=`); synchronized uPlot crosshairs, brush-to-zoom, compare-to-previous overlay, and shared event annotations. |
| 6 | **Live by default, pausable** | Space/toggle sets `$live=false`; server patches only `#pending` counter while paused and sends one full snapshot on resume. Local 1 s JS ticker flags widgets stale when `age > 2 × cadence`. |
| 7 | **Color-blind-safe status semantics** | Fixed 5-state palette (`ok`, `warn`, `crit`, `info`, `unknown`) always paired with distinct icon shape + text label + dashed chart thresholds. |
| 8 | **Budgets & offline WCAG 2.2 AA** | First paint `< 300 ms` LAN, payload `< 150 KB` brotli, CPU `< 3%` of 1× A72 core, RSS `< 40 MB`, flash writes `< 50 MB/day`, 0 serious `axe-core` violations, zero CDN/internet calls. |

---

## 3. Collector Cadence, Privilege & CPU Cost Matrix

Distilled from Brief §9.3, §12.1, and Appendix A (estimates on 1× Cortex-A72 core):

| Source | Mechanism | Privilege | Cadence | Est. Cost/Run | % of 1 Core | Retention |
| --- | --- | --- | --- | --- | --- | --- |
| CPU, load, memory | `/proc/stat`, `loadavg`, `meminfo` | Unpriv | `1 s` / `5 s` | `0.12 ms` | `0.012%` | T0–T4 |
| Link stats (12 ifaces) | Netlink `RTM_GETLINK` stats64 | Unpriv | `1 s` | `0.25 ms` | `0.025%` | T0–T4 |
| IRQ / softirq | `/proc/softirqs`, `/proc/interrupts` | Unpriv | `5 s` | `0.30 ms` | `0.006%` | T1–T4 |
| Thermal / hwmon | `/sys/class/hwmon/*`, `thermal/*` | Unpriv | `5 s` | `0.15 ms` | `0.003%` | T1–T4 |
| Conntrack stats | Netlink `CTNETLINK` stats | `CAP_NET_ADMIN` | `5 s` | `0.20 ms` | `0.004%` | T0–T4 |
| nftables counters (120 rules) | Netlink `NFT` dump | `CAP_NET_ADMIN` | `5 s` | `2.50 ms` | `0.050%` | T2–T4 |
| WireGuard peers | Netlink `wgctrl` | `CAP_NET_ADMIN` | `5 s` | `0.30 ms` | `0.006%` | T2–T4 |
| QoS qdisc / CAKE xstats | Netlink `RTM_GETQDISC/TCLASS` | Unpriv | `5 s` | `0.60 ms` | `0.012%` | T1–T4 |
| NIC driver stats (5 ports) | Ethtool netlink / ioctl | Unpriv* | `5 s` | `0.80 ms` | `0.016%` | Pinned T2–T4 |
| Per-client accounting | `CTNETLINK` dump (5k flows) | `CAP_NET_ADMIN` | `10 s` | `18.0 ms` | `0.180%` | Top-50 T2–T4 |
| Conntrack full table | `CTNETLINK` dump (50k flows) | `CAP_NET_ADMIN` | `5 s` *(open only)* | `~150 ms` | *(+3.0% while open)* | Live only |
| FRR (BGP/OSPF/RIB/BFD) | `vtysh -c '... json'` (batched) | `frrvty` group | `30 s` (`5 s` BFD) | `45.0 ms` | `0.150%` | T2–T4 + events |
| Kea DHCP leases & pools | UNIX control socket JSON | Socket read | `30 s` | `6.0 ms` | `0.020%` | T2–T4 |
| PowerDNS Recursor | `rec_control get-all` | Allowlisted exec | `30 s` | `12.0 ms` | `0.040%` | T2–T4 |
| Chrony NTP | `chronyc -c sources/tracking` | Allowlisted exec | `60 s` | `15.0 ms` | `0.025%` | T2–T4 |
| LLDP / systemd units | `lldpcli -f json` / D-Bus `ListUnits` | Unpriv | `60 s` / `30 s` | low | `< 0.01%` | Events |
| VyOS op-mode (structured) | Allowlisted `vyos.opmode` raw/JSON | Unpriv / `vyattacfg-ro` | `300 s` / commit | `650.0 ms` | `0.217%` | Archive / T2–T4 |
| SFP+ DOM optics (2 cages) | Ethtool module EEPROM (I2C) | `CAP_NET_ADMIN` | `300 s` | `40.0 ms` | `0.013%` | T2–T4 |
| eMMC wear (`EXT_CSD`) | `/sys/class/mmc_host/*/life_time` | Unpriv | `3600 s` | `30.0 ms` | `0.001%` | T4 |
| Running config & archive | `/config/archive` + `inotify` | `vyattacfg-ro` | On commit | `25.0 ms` | `0.001%` | Archive |
| journald log stream | `journalctl -o json -f` | `systemd-journal` | Event-driven | `0.05 ms` | `0.005%` | Journal + 30 d hist |
| Rollup + flush + SSE render | In-process | Unpriv | `60 s` / `1 s` | `25.0` / `1.6 ms` | `0.042%` + `0.160%` | — |
| **Total (idle-streaming, 1 client)** | | | | | **`~0.99%` of 1 core** | Budget: `< 3%` |

---

## 4. Security, Redaction & Licensing Checklist

### 4.1 Security Hardening (Brief §11)
- **Two-process split (same binary):** `vyview web` holds zero capabilities + `NoNewPrivileges=yes`, `ProtectSystem=strict`, `SystemCallFilter=@system-service`, `MemoryMax=96M`, `CPUQuota=25%`, `GOMEMLIMIT=40MiB`, `GOMAXPROCS=2`. `vyview probe` holds only `CAP_NET_ADMIN` with `IPAddressDeny=any`.
- **Sessions & rate limiting:** 256-bit random session IDs in RAM (12 h idle / 7 d absolute, rotated on login), `Secure; HttpOnly; SameSite=Strict` cookies; rate limits (login 5/min/IP with lockout, search 10/s, export 1/s, new streams 10/min).
- **Server-side secret redaction (Brief §11.8):** Redacted on load *before* indexing, rendering, search, or diffing across 19 secret config path patterns (VyOS API keys, GraphQL JWT secrets, IPsec PSKs/keys, WireGuard private keys/PSKs, login password hashes, SSH public keys, RADIUS/TACACS keys, PKI private keys, BGP/OSPF/SNMP/PPPoE/Wi-Fi/DDNS/OpenVPN/VRRP secrets, and heuristic container env vars). Replaced with a constant marker + per-boot HMAC hash so diffs show `•••••• (changed)` with **no reveal control**.
- **Household privacy (Brief §11.9):** `privacy.per_client = off | aggregate | full`; top-N only by default; 30 d client detail retention; no DNS query logging.
- **Datastar CSP mitigation (Brief §11.6):** All Datastar `data-*` expressions are static template literals; a CI template lint fails on any interpolated/dynamic `data-*` attribute.

### 4.2 Font Awesome Pro 7.3.1 Dual-Pack Licensing (Brief §9.9, §9.11)
- **Version pin:** Exactly `7.3.1` in lockfile; Renovate updates gated on manual review; `icons.provenance.json` emitted per build.
- **Dual icon pack with identical symbol IDs:**
  - Private builds (`go build -tags fapro` / `VYVIEW_ICON_PACK=pro`): Font Awesome Pro 7.3.1 (**Sharp Light** `fa-sharp fa-light` for navigation, status, KPI, and chrome). Source: Font Awesome CLI (`fa`) / downloaded "By Icon" Kit or `@fortawesome/sharp-light-svg-icons` via CI secret token.
  - Public / upstream VyOS builds (default): Font Awesome Free 7.x fallback sprite mapping every Pro-only glyph (`router`, `chart-network`, `shield-check`, `block-brick-fire`, `list-timeline`, `lock-keyhole`, `filter-list`, `gauge-max`, `wave-pulse`, `signal-bars`, `command`, `radar`) 1:1 onto Free glyphs.
- **Leak guard:** `.gitignore` blocks `icons/pro/`, `icons.provenance.json`, `.npmrc`; CI fails any public build referencing Pro SVG path hashes, `@fortawesome/pro-*`, or `-tags fapro`.

---

## 5. Key Findings, Architectural Refinements & Open Decisions

### Architectural Refinements Applied to This Repo
1. **Single-file multi-ring datastore (`vyview.ring`):** Consolidates the Brief's separate per-tier RRD files + separate event log file + optional T1 shutdown dump into **one pre-allocated, 4 KiB-aligned binary file** (see [STORAGE-FORMAT.md](STORAGE-FORMAT.md)), reducing ext4 metadata churn and matching the strict **1 binary + 1 data file** router footprint.
2. **Static binary (`CGO_ENABLED=0`) friction points flagged early:**
   - Use `journalctl -o json -f` instead of cgo `sd-journal`.
   - Resolve PAM authentication (`libpam` requires cgo) before MVP v0.1: either pair pure-Go auth (`argon2id` + mTLS + forward-auth) with a narrow allowlisted OS PAM verifier, or adopt an `arm64` cross-CGO build pipeline if in-process PAM is mandatory.

### Open Decisions to Close in MVP v0.1 (Brief §15.2)
| Decision | Options | Recommendation |
| --- | --- | --- |
| **Template engine** | `templ` vs stdlib `html/template` | Prototype both in Week 1; prefer `templ` for compile-time type safety and fast zero-reflection SSE fragment renders. |
| **PAM auth vs pure `CGO_ENABLED=0`** | Subprocess PAM helper / pure-Go local+mTLS vs cross-CGO `libpam` | Keep `vyview` `CGO_ENABLED=0` via pure-Go `argon2id`/mTLS + narrow PAM helper (or validate cross-CGO early). |
| **Chrome icon family** | Classic Regular (default) vs Sharp Regular | Review both on dark/light themes in first visual pass; never mix Classic and Sharp on the same surface. |
| **ASK2 adapter contract** | `ask2ctl --json` + DPAA1 sysfs/debugfs schema | Freeze capability & JSON schema early; ship stubbed fake adapter for v0.1–v0.3, wire real ASK2 in v0.4. |
| **Packaging sequence** | (a) `.deb` in custom VyOS image → (c) upstream `set service dashboard` | Start with (a) for the LS1046A Mono Gateway image; propose (c) upstream after v0.2 stabilizes. |

### Roadmap Summary (Brief §15)
- **MVP v0.1 (6–8 pw):** App shell, auth, Overview, Interfaces (list + ethernet detail), System, Configuration viewer (tree/set/diff/history + redaction), Logs (tail), T0/T1/T2 single-file ring store, SSE hub, `vyview-probe` helper.
- **v0.2 (5–7 pw):** Firewall/NAT counters, virtualized Conntrack table (50k+ rows), Kea/DNS/NTP services, WireGuard & IPsec VPN, full T2–T4 flash retention + `< 50 MB/day` measurement, CSV/JSON export.
- **v0.3 (5–7 pw):** FRR Routing (BGP/OSPF/BFD/RIB/FIB/VRFs), QoS/CAKE, Topology map, Ctrl-K command palette & global inverted index (`< 50 ms` p95), Health & Alerts + chart annotations, scheduled probes.
- **v0.4 (4–6 pw):** ASK2 Dataplane module (with stock x86 graceful degradation), lazy ECharts Sankey/heatmaps, kiosk mode, installable PWA, mobile polish.
