# Roadmap, Open Decisions & Risks

> Milestone checklists from MVP `v0.1` through `v0.4` and beyond, open architectural decisions to close early, and tracked technical risks with mitigations.

*(Effort estimates assume one experienced engineer: `~24–36` person-weeks total to `v0.4`.)*

---

## 1. Release Milestones

### MVP `v0.1` — Foundation & Core Router Visibility (`6–8 pw`)
- [ ] **Read-only skeleton & CI:** Go module layout, declarative `GET`/`HEAD` data-plane mux (+ `/login`/`/logout` POST exceptions), CSP + static Datastar expression lint, allowlisted `internal/exec` runner + static `os/exec` check, `arm64`/`amd64` cross-compile, `govulncheck`, and `< 150 KB` payload gate.
- [ ] **Authentication & app shell:** HTTPS-by-default LAN listener, PAM / VyOS operator-group / `argon2id` auth, collapsible left activity rail, breadcrumbs, health pill, theme/density switcher.
- [ ] **Core pages:** Overview (plain-language health, WAN card, 7 KPIs, linked talkers/events), Interfaces (3-layer view + Ethernet detail), System (CPU heatmap, TMP431 thermals, translated eMMC wear, boot images), Configuration (collapsible tree, single `set` block, red/green two-revision diff, audited commit history, 19-class secret redaction), Logs & Events (severity-stacked histogram + journal tail).
- [ ] **Live data & storage core:** `200 ms` `T0` RAM slab, `5 s` `T1` `tmpfs` ring, `60 s` `T2` ring inside `vyview.ring`, SSE hub with backpressure, and `vyview-probe` privileged helper skeleton.
- [ ] **Exit criteria:** LS1046A resource budgets met (`< 3%` CPU, `< 40 MB` RSS, `< 300 ms` warm paint), golden parser tests for every used command, `0` serious/critical `axe-core` violations.

### `v0.2` — Traffic, Security, VPN & Full Flash Retention (`5–7 pw`)
- [ ] **Firewall & NAT:** Zone policy matrix above dense FortiGate/OPNsense-style grouped rule table, **Firewall Security Check** (*"Is my firewall safe?"*), groups, NAT/NAT66, virtualized `50k+` Conntrack table with offload path & *"why not offloaded"* reasons.
- [ ] **Traffic & Clients:** Two-plane per-client Top-K accounting (`CTNETLINK` + offload reconciler), stacked top-6 + other chart, client/OT asset inventory.
- [ ] **VPN & Services:** WireGuard (`wgctrl` + `wgfp_rx`), IPsec (VICI + `xfrm` + CAAM), OpenVPN; Services with **Podman containers** (`podman.sock`), Kea DHCP, PowerDNS, chrony/PTP, and explicit per-daemon control-socket health.
- [ ] **Storage & export:** Full `T2`–`T4` retention (`14 d` / `90 d` / `2 y`) + Compact profile + `--data-path` USB + `--snapshot-path`, bounded `≤ 4 MiB` T2+ warming cache, CSV/JSON/PNG export, recommended Prometheus `/metrics` (incl. health & alerts).
- [ ] **Exit criteria:** `50k` conntrack load test passes within budgets; 7-day hardware flash-write measurement `< 50 MB/day`.

### `v0.3` — Routing, QoS, Redundancy, Topology & Search (`5–7 pw`)
- [ ] **Routing:** FRR JSON + netlink with VRF selector & per-VRF RIB/FIB tabs, longest-prefix lookup, BGP/OSPF/BFD/PBR, churn.
- [ ] **QoS:** Ingress & egress CAKE/FQ-CoDel shapers, all tins (Voice/Video delay warnings), ECN marks vs. drops, offload-bypass share, bufferbloat grade.
- [ ] **Redundancy:** Capability-detected failover hub (`VRRP` via `keepalived` D-Bus, `conntrack-sync`, WAN failover, `MRP` ring events once supported in VyOS).
- [ ] **Topology & Diagnostics:** Layered L2/L3 map (LLDP switches/APs, zone shading, negotiated speed colors, collapsible client groups), step-by-step WAN connectivity chain (*"Why is my internet not working?"*), scheduled probe trends.
- [ ] **Global search & Health:** `Ctrl/Cmd-K` command palette (CLI-aware, `< 50 ms` p95), **Read-Only Security Check** (*"Am I protected well enough?"*), chart event annotations.
- [ ] **Exit criteria:** Palette `< 50 ms` p95; topology verified on reference and lab configs; propose upstream `service dashboard` config node after v0.2/v0.3 stabilization.

### `v0.4` — Unified Acceleration, Opt-In MCP & Polish (`4–6 pw`)
- [ ] **Dataplane / Acceleration:** Unified capability-detected tree (`Summary`, `Forwarding: ASK2 / VPP / Flowtable / AF_XDP`, `Crypto: CAAM / QAT`, `Other`), preferring **ASK2 Generic Netlink (`ynl --family ask` / `show offload flow`)** over legacy debugfs, plus OT/TSN L2 protocol visibility.
- [ ] **Opt-in MCP interface:** Read queries answering every job in 1 prompt, gated on-demand diagnostics (`ping`/`traceroute`), and opt-in configuration writes strictly via `set → commit-confirm → save` with `100%` audit logging and automated write-safety tests.
- [ ] **Visual & mobile polish:** Lazy ECharts Sankey (WAN↔segment + inter-VLAN) and large heatmaps, multi-vendor CLI cheat sheets, kiosk rotation, installable PWA, mobile polish.
- [ ] **Exit criteria:** Full Playwright journey parity on both stock `x86_64` (graceful degradation) and `LS1046A` (ASK2 enabled); OT/TSN forwarding latency no-harm verified on hardware.

### Post-`v0.4` (Requires Separate Threat Model per Feature)
- Remote/off-router polling mode (Section 8 of [System Architecture](architecture/system-architecture.md)), read-only multi-router federation, remote-write TSDB export, and optional time-boxed packet diagnostics.

---

## 2. Open Architectural Decisions to Close in MVP Week 1

| Decision | Options | Recommendation |
| --- | --- | --- |
| **Template engine** | `templ` vs. stdlib `html/template` | Time-boxed Week 1 spike; prefer `templ` for compile-time type safety and fast zero-reflection SSE fragment rendering. |
| **PAM auth vs. `CGO_ENABLED=0`** | Pure-Go (`argon2id` + mTLS + forward-auth) + narrow PAM helper vs. cross-CGO `libpam` | Keep vyview `CGO_ENABLED=0` via pure-Go auth + narrow OS PAM verifier subprocess; validate against the VyOS user/operator-group model early. |
| **Datastar CSP requirement** | Verify whether pinned Datastar v1.0.x still requires `script-src 'unsafe-eval'` | Check the pinned SDK release in Week 1; retain the static `data-*` template lint regardless. |
| **ASK2 Generic Netlink schema** | `ynl --family ask` / `show offload flow` JSON & netlink attribute contract | Define capability & attribute contract with `we-are-mono/ASK`; ship fake stub for `x86_64` and CI from Day 1. |

---

## 3. Risk Register

| Risk | Mitigation | Owner |
| --- | --- | --- |
| **VyOS op-mode / FRR JSON drift across releases** | Prefer kernel netlink, `/proc`, `/sys`, and local sockets; versioned tolerant parsers with weekly rolling golden captures. | Dev |
| **Datastar `'unsafe-eval'` CSP surface** | Static `data-*` template lint (CI fail on dynamic attributes), contextual auto-escaping, track upstream CSP-safe Datastar builds. | Dev / Sec |
| **eMMC flash wear on embedded routers** | Time-major `vyview.ring` (`< 50 MB/day`), RAM `T0`/`T1`, Compact profile (`< 5 MB/day`), `--data-path` USB support, continuous `/proc/self/io` & `EXT_CSD` wear self-monitoring (`> 50%` advisory). | Arch |
| **`CAP_NET_ADMIN` privilege surface** | Isolated `vyview-probe` unit (`IPAddressDeny=any`), fixed read-only RPC & netlink `*_GET*` allowlist, fuzzed decoders. | Arch / Sec |
| **ASK2 ABI churn (debugfs vs. netlink)** | Prefer Generic Netlink (`ynl --family ask`) over `fe_flow` debugfs; keep debugfs behind defensive schema/version guards inside `vyview-probe`. | Arch / ASK2 |
| **BMan buffer-pool exhaustion (512-entry pool)** | Surface free % (`< 30%` warn / `< 15%` crit), correlate with source-split drops, document tuning guidance. | ASK2 |
| **Household / OT client privacy** | Conservative defaults (top-N + `other`, `30 d` detail, zero DNS query logging), `privacy.per_client = off \| aggregate \| full`. | Owner |
| **Upstream VyOS config-node acceptance** | Ship `.deb` in custom LS1046A image first (Option a); propose `set service dashboard` upstream after v0.2 proves stable. | Owner |
