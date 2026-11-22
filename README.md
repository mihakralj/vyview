# vyview

**vyview** is a read-only, real-time web dashboard for [VyOS](https://vyos.io) routers. It provides fast, visual access to everything VyOS op-mode and the underlying Linux kernel already know — interface and client traffic, firewall and NAT counters, FRR routing state, VPN health, network services and Podman containers, QoS/CAKE shaping, redundancy/failover state, hardware/flash health, redacted configuration and commit diffs, journal logs, and capability-detected hardware-offload telemetry (ASK2, VPP, Flowtable, AF_XDP, CAAM, QAT).

It is **read-only by construction in the web UI** (no mutating HTTP endpoints, no write collectors, no shell execution, and no VyOS HTTP API key required in on-box mode). Alongside the web GUI, an **opt-in MCP interface** lets AI assistants answer operational questions and apply audit-logged configuration changes exclusively through VyOS's native `set → commit-confirm → save` workflow.

## Deployment Footprint: Two Files on the Router

`vyview` is designed for embedded ARM64 routers (reference target: **NXP LS1046A "Mono Gateway"**, 4× Cortex-A72, Linux 6.18, VyOS 1.5/rolling) as well as stock `x86_64` VyOS — with **zero external runtime dependencies** (no Node.js, no Python web server, no external database):

| File on Router | Role |
| --- | --- |
| **1. `/usr/bin/vyview`** | Single static Go binary (`go:embed`ded HTML templates, UnoCSS, Font Awesome SVG sprite, Datastar ES module, uPlot, lazy ECharts, WOFF2 fonts, offline docs). Invoked as two hardened systemd units from the same binary: `vyview web` (unprivileged, zero capabilities) and `vyview probe` (`CAP_NET_ADMIN` read-only helper over a UNIX socket). |
| **2. `/var/lib/vyview/vyview.ring`** | Single pre-allocated, fixed-size **multi-ring binary datastore** (configurable via `--data-path` for external USB) holding the series catalog, time-major rollup rings (`60 s × 14 d`, `5 min × 90 d`, `1 h × 2 y` — or compact `5 min × 7 d`, `1 h × 1 y`), circular event/alert/probe ring, and an optional clean-shutdown snapshot of the in-RAM `5 s` hot ring. |

Live `200 ms` (5 Hz) rings (`~1.5 MiB`, `60 fps` interpolated in `<vy-chart>`) and the `5 s` hot tier (`tmpfs`) live strictly in RAM to protect eMMC/SD flash endurance while keeping live charts fluid.

## Core Technology Stack

- **Backend & Collectors:** Go (`CGO_ENABLED=0` static cross-compile for `linux/arm64` & `linux/amd64`), netlink (`vishvananda/netlink`, `mdlayher/netlink`, `google/nftables`, `ti-mo/conntrack`, `florianl/go-tc`, `wgctrl`, `mdlayher/ethtool`, ASK2 Generic Netlink `ynl --family ask`), direct `/proc` & `/sys` readers (including kernel sysfs `EXT_CSD` eMMC wear with zero `mmc-utils` dependency), local UNIX sockets (Kea, VICI, Podman), and allowlisted read-only CLIs (`vtysh`, `journalctl`, `chronyc`, `lldpcli`, `rec_control`).
- **Storage:** Custom single-file, page-aligned, time-major multi-ring datastore (`fdatasync` batched every `60 s`, per-slot CRC32C, statistical cross-tier rollups, bounded `≤ 4 MiB` RAM tier warming, zero compaction, zero growth).
- **Hypermedia & Real-Time UI:** [Datastar](https://data-star.dev) v1.0.x (~11 KB brotli self-hosted ES module) + official Datastar Go SDK (`github.com/starfederation/datastar-go/datastar`) over a single Server-Sent Events (`text/event-stream`) connection per open page.
- **Charts, CSS & Icons (build-time only):** [uPlot](https://github.com/leeoniya/uPlot) (~19.5 KB brotli) wrapped in a `<vy-chart>` Web Component, server-rendered inline SVG (sparklines, small heatmaps, gauges, topology), lazy-loaded modular Apache ECharts (Sankey/large heatmaps), [UnoCSS](https://unocss.dev) CLI (`app.css`, ~6 KB brotli), locally embedded **IBM Plex Sans** & **IBM Plex Mono** WOFF2 fonts, and a subsetted SVG `<symbol>` sprite (Font Awesome Pro 7.3.1 **Sharp Light** for private builds via `-tags fapro`; Font Awesome Free 7.x fallback for public/upstream builds).

## Target Resource Budgets (LS1046A Reference)

| Budget | Target | Estimated |
| --- | --- | --- |
| Idle-streaming CPU (1 Overview client) | `< 3%` of 1× A72 core | `~0.99%` |
| Resident memory (`VmRSS`) | `< 40 MB` (`GOMEMLIMIT=40MiB`) | `~30.1 MB` |
| Flash writes (host) | `< 50 MB/day` | `~28–32 MB/day` full / `< 5 MB/day` compact |
| First contentful paint (LAN, warm / cold) | `< 300 ms` / `< 600 ms` | Initial payload `~140 KB` brotli (`~70 KB` system fonts) |
| Counter change to DOM latency | `< 1.5 s` | `200 ms` fast loop + `60 fps` interpolation |
| Accessibility | WCAG 2.2 AA | `0` serious/critical `axe-core` violations |

## Local Frontend Demo

The dashboard prototype uses deterministic **demo data only**. It does not read router data, connect to a router, or implement authentication. It is not suitable for production deployment or internet exposure. By default the server binds only to loopback addresses (`-allow-lan` permits LAN evaluation on a test router).

From the repository root, start the demo with Go 1.26.11 or newer. The bundled CSS is already available, so this command does not require mise or npm:

```sh
go run ./cmd/vyview
```

Open `http://127.0.0.1:8080`. Time ranges are fixture-backed and retained in the URL. Switch scenarios with `?scenario=normal`, `?scenario=warning` (default), or `?scenario=unavailable`. For a different local port, run `go run ./cmd/vyview -addr 127.0.0.1:8081`.

When developing the UI, `mise exec -- make demo` rebuilds the CSS before starting the server. With Go, Node, and npm already on your PATH, use `npm ci` to install dependencies and `make demo` to rebuild and run without mise.

Use the sidebar (collapsible via `[`) or **More** on mobile to visit all demo pages. Press **Ctrl+K** or **Cmd+K** to search pages and fixture records. Try an interface detail tab, longest-prefix lookup under **Routing → RIB / FIB**, the **Configuration → Diff** inspector, the clickable topology, or the simulated journal replay. **Settings** changes the scenario and theme; these preferences are carried in links rather than stored globally.

The demo defaults to the dark palette from the [Design System](docs/ui/design-system.md). In **Settings**, choose **Light theme** or **System theme** to use the light palette or follow your system preference.

The interface uses locally embedded [IBM Plex Sans 1.3.6](https://github.com/IBM/plex/releases/tag/%40ibm%2Fplex-sans%401.3.6) and [IBM Plex Mono 2.5.0](https://github.com/IBM/plex/releases/tag/%40ibm%2Fplex-mono%402.5.0). Both are licensed under the [SIL Open Font License 1.1](internal/demo/assets/fonts/LICENSE.txt). No font requests are sent to external services.

The [Mockup Index](docs/ui/dashboard-features.md#mockups) is the visual screen reference. The demo covers those destinations and their main drill-down paths, but does not reproduce every widget or control. Live SSE subscriptions, production configuration handling, login, and privileged collection are not part of this demo.

### Replacing Demo Data

The HTTP handler accepts a read-only [`Provider`](internal/demo/pages.go) via `HandlerWithProvider`. The default [`Fixtures`](internal/demo/fixtures.go) implementation returns in-memory snapshots; the HTTP layer owns navigation, filtering, sorting, and presentation. A router-backed adapter can implement the same contract, honor the selected historical range, return `ErrNotFound` for absent entities, and return no records or samples for unavailable sources. Sanitize or redact sensitive data before constructing snapshots.

## Project Tooling

Install [mise](https://mise.jdx.dev/getting-started.html) as a user-accessible executable, then from the repository root:

```sh
export MISE_DATA_DIR="$PWD/.tools/mise"
export MISE_STATE_DIR="$PWD/.tools/mise-state"
export MISE_CACHE_DIR="$PWD/.tools/mise-cache"
mise trust
mise install
mise exec -- npm ci
mise exec -- make check
```

`mise.toml` pins Go and Node; npm and Go dependencies are locked in `package-lock.json`, `go.mod`, and `go.sum`. No router-side Node installation is needed.

To run the browser and accessibility smoke test, install the pinned Playwright Chromium into the project and run it:

```sh
export PLAYWRIGHT_BROWSERS_PATH="$PWD/.tools/playwright"
mise exec -- npx playwright install chromium
mise exec -- make browser
```

`make check` builds the demo's UnoCSS utilities and tests static Go cross-compilation for Linux `arm64` and `amd64`. `mise exec -- make security` runs `govulncheck` and `npm audit`; `mise exec -- make sbom` writes CycloneDX inventories under `dist/`.

## Documentation

All architecture specifications, UI standards, page specs, roadmap milestones, review logs, and original reference mockups are organized under **[docs/README.md](docs/README.md)**:

- **[Product Overview](docs/architecture/overview.md)** — Vision, five UX principles, personas (`P1`–`P6`), question workflows, and success metrics.
- **[System Architecture](docs/architecture/system-architecture.md)** — Two-process runtime split, Two-Plane Offload Model, collector catalogue, Datastar/SSE, and Local vs. Remote differences.
- **[Storage Format (`vyview.ring`)](docs/architecture/storage-format.md)** — Binary multi-ring specification (`200 ms`–`2 y`), flash-wear math, statistical rollups, and crash recovery.
- **[Security & MCP Model](docs/architecture/security-and-mcp.md)** — Read-only GUI enforcement, secret redaction, privacy, FA Pro dual-pack licensing, and opt-in MCP `commit-confirm` safety.
- **[Deployment & Packaging](docs/architecture/deployment-and-packaging.md)** — `.deb` / upstream CLI sketch, RBAC `operator-group`, systemd units, USB `--data-path`, and delayed boot start.
- **[Performance & Testing](docs/architecture/performance-and-testing.md)** — LS1046A resource budgets, per-collector CPU estimates, and CI/soak/MCP safety gates.
- **[UI & Page Specifications](docs/ui/information-architecture.md)** — [Design System](docs/ui/design-system.md), [Dashboard Features & Mockups](docs/ui/dashboard-features.md), and all [18 Page Specifications](docs/ui/pages/README.md).
- **[Roadmap, Decisions & Risks](docs/roadmap.md)** — `v0.1`–`v0.4` milestones, open Week 1 decisions, and risk register.
- **[Reviews & Audits](docs/reviews/monitoring-gap-analysis.md)** — [Design Brief Analysis](docs/reviews/design-brief-analysis.md), [Monitoring Gap Analysis](docs/reviews/monitoring-gap-analysis.md), [@Haico Feedback](docs/reviews/feedback-haico.md), and [@Daegara Feedback](docs/reviews/feedback-daegara.md).
