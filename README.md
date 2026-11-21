# vyview

**vyview** is a read-only, real-time web dashboard for [VyOS](https://vyos.io) routers. It provides fast, visual access to everything VyOS op-mode and the underlying Linux kernel already know — interface and client traffic, firewall and NAT counters, FRR routing state, VPN health, network services, QoS/CAKE shaping, hardware/flash health, redacted configuration and commit diffs, journal logs, and optional DPAA1/ASK2 hardware-offload telemetry.

It is **read-only by construction**: there are no mutating HTTP endpoints, no write collectors, no shell execution, and no VyOS HTTP API key required in on-box mode.

## Deployment Footprint: Two Files on the Router

`vyview` is designed for embedded ARM64 routers (reference target: **NXP LS1046A "Mono Gateway"**, 4× Cortex-A72, Linux 6.18, VyOS 1.5/rolling) as well as stock x86_64 VyOS — with **zero external runtime dependencies** (no Node.js, no Python web server, no external database):

| File on Router | Role |
| --- | --- |
| **1. `/usr/bin/vyview`** | Single static Go binary (`go:embed`ded HTML templates, UnoCSS, Font Awesome SVG sprite, Datastar ES module, uPlot, lazy ECharts, WOFF2 fonts, offline docs). Invoked as two hardened systemd units from the same binary: `vyview web` (unprivileged, zero capabilities) and `vyview probe` (`CAP_NET_ADMIN` read-only helper over a UNIX socket). |
| **2. `/var/lib/vyview/vyview.ring`** | Single pre-allocated, fixed-size **multi-ring binary datastore** holding the series catalog, time-major rollup rings (1 min × 14 d, 5 min × 90 d, 1 h × 2 y), circular event/alert/probe ring, and an optional clean-shutdown snapshot of the in-RAM 10 s hot ring. |

Live 1-second rings (`0.5 MB`) and the 10-second hot tier (`41.7 MB`) live strictly in RAM (`process` / `tmpfs`) to protect eMMC/SD flash endurance.

## Core Technology Stack

- **Backend & Collectors:** Go (`CGO_ENABLED=0` static cross-compile for `linux/arm64` & `linux/amd64`), netlink (`vishvananda/netlink`, `mdlayher/netlink`, `google/nftables`, `ti-mo/conntrack`, `florianl/go-tc`, `wgctrl`, `mdlayher/ethtool`), direct `/proc` & `/sys` readers, local UNIX sockets (Kea, VICI), and allowlisted read-only CLIs (`vtysh`, `journalctl`, `chronyc`, `lldpcli`, `rec_control`).
- **Storage:** Custom single-file, page-aligned, time-major multi-ring datastore (`fdatasync` batched every 60 s, per-slot CRC32C, zero compaction, zero growth).
- **Hypermedia & Real-Time UI:** [Datastar](https://data-star.dev) v1.0.x (~11 KB brotli self-hosted ES module) + official Datastar Go SDK (`github.com/starfederation/datastar-go/datastar`) over a single Server-Sent Events (`text/event-stream`) connection per open page.
- **Charts, CSS & Icons (build-time only):** [uPlot](https://github.com/leeoniya/uPlot) (~19.5 KB brotli) wrapped in a `<vy-chart>` Web Component, server-rendered inline SVG (sparklines, small heatmaps, gauges, topology), lazy-loaded modular Apache ECharts (Sankey/large heatmaps), [UnoCSS](https://unocss.dev) CLI (`app.css`, ~6 KB brotli), and a subsetted SVG `<symbol>` sprite (Font Awesome Pro 7.3.1 for private builds via `-tags fapro`; Font Awesome Free 7.x fallback for public/upstream builds).

## Target Resource Budgets (LS1046A Reference)

| Budget | Target | Estimated |
| --- | --- | --- |
| Idle-streaming CPU (1 Overview client) | `< 3%` of 1× A72 core | `~0.99%` |
| Resident memory (`VmRSS`) | `< 40 MB` (`GOMEMLIMIT=40MiB`) | `~30.1 MB` |
| Flash writes (host) | `< 50 MB/day` | `~21–32 MB/day` (`< 0.71%` eMMC endurance / 10 y) |
| First contentful paint (LAN, warm / cold) | `< 300 ms` / `< 600 ms` | Initial payload `~140 KB` brotli (`~70 KB` system fonts) |
| Counter change to DOM latency | `< 1.5 s` at 1 s cadence | `< 1.1 s` |
| Accessibility | WCAG 2.2 AA | `0` serious/critical `axe-core` violations |

## Documentation

- [docs/TECH-STACK.md](docs/TECH-STACK.md) — Complete technology list required on the router (gather, store, Datastar presentation), OS/kernel prerequisites, `CGO_ENABLED=0` trade-offs, and CI-only build toolchain.
- [docs/STORAGE-FORMAT.md](docs/STORAGE-FORMAT.md) — Binary layout and engineering specification for the single-file multi-ring datastore (`vyview.ring`).
- [docs/DESIGN-BRIEF-ANALYSIS.md](docs/DESIGN-BRIEF-ANALYSIS.md) — Engineering analysis of the *VyOS View-Only Dashboard Design Brief v1.1 (FA Pro)*, collector cadence matrix, security checklist, licensing rules, and open architectural decisions.
