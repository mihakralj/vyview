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
- **Charts, CSS & Icons (build-time only):** [uPlot](https://github.com/leeoniya/uPlot) (~19.5 KB brotli) wrapped in a `<vy-chart>` Web Component, server-rendered inline SVG (sparklines, small heatmaps, gauges, topology), lazy-loaded modular Apache ECharts (Sankey/large heatmaps), [UnoCSS](https://unocss.dev) CLI (`app.css`, ~6 KB brotli), locally embedded **IBM Plex Sans** & **IBM Plex Mono** WOFF2 fonts, and a subsetted SVG `<symbol>` sprite (Font Awesome Pro 7.3.1 **Sharp Light** for private builds via `-tags fapro`; Font Awesome Free 7.x fallback for public/upstream builds).

## Target Resource Budgets (LS1046A Reference)

| Budget | Target | Estimated |
| --- | --- | --- |
| Idle-streaming CPU (1 Overview client) | `< 3%` of 1× A72 core | `~0.99%` |
| Resident memory (`VmRSS`) | `< 40 MB` (`GOMEMLIMIT=40MiB`) | `~30.1 MB` |
| Flash writes (host) | `< 50 MB/day` | `~21–32 MB/day` (`< 0.71%` eMMC endurance / 10 y) |
| First contentful paint (LAN, warm / cold) | `< 300 ms` / `< 600 ms` | Initial payload `~140 KB` brotli (`~70 KB` system fonts) |
| Counter change to DOM latency | `< 1.5 s` at 1 s cadence | `< 1.1 s` |
| Accessibility | WCAG 2.2 AA | `0` serious/critical `axe-core` violations |

## Local Frontend Demo

The dashboard prototype uses deterministic **demo data only**. It does not read router data, connect to a router, or implement authentication. It is not suitable for deployment on a router or exposure to a network. The server only accepts loopback IP addresses.

From the repository root, start the demo with Go 1.26.11 or newer. The bundled CSS is already available, so this command does not require mise or npm:

```sh
go run ./cmd/vyview
```

Open `http://127.0.0.1:8080`. Time ranges are fixture-backed and retained in the URL. The default warning scenario can be changed to `?scenario=normal` or `?scenario=unavailable` to inspect other states. The app serves embedded assets and does not require network access once built. For a different local port, run `go run ./cmd/vyview -addr 127.0.0.1:8081`.

When developing the UI, `mise exec -- make demo` rebuilds the CSS before starting the server. With Go, Node, and npm already on your PATH, use `npm ci` to install dependencies and `make demo` to rebuild and run without mise.

Use the sidebar or **More** on mobile to visit all demo pages. Press **Ctrl+K** or **Cmd+K** to find pages and fixture records. Try an interface detail tab, longest-prefix lookup under **Routing → RIB / FIB**, the **Configuration → Diff** inspector, the clickable topology, or the simulated journal replay. **Settings** changes the scenario and theme; these preferences are carried in links rather than stored globally.

The demo defaults to the dark palette from the [design system](plans/03-design-system.md). In **Settings**, choose **Light theme** or **System theme** to use the light palette or follow your system preference.

The interface uses locally embedded [IBM Plex Sans 1.3.6](https://github.com/IBM/plex/releases/tag/%40ibm%2Fplex-sans%401.3.6) and [IBM Plex Mono 2.5.0](https://github.com/IBM/plex/releases/tag/%40ibm%2Fplex-mono%402.5.0). Both are licensed under the [SIL Open Font License 1.1](internal/demo/assets/fonts/LICENSE.txt). No font requests are sent to external services.

The [mockup index](plans/04-dashboard-features.md#mockups) is the screen reference. The demo covers those destinations and their main drill-down paths, but does not reproduce every widget or control. Advanced charts, live subscriptions, production configuration handling, login, and privileged collection are not part of this demo.

The Interfaces inventory has fixture status counts, multi-select type filters, sortable and configurable columns, address copy, error indicators, and example one-hour trends. Column preferences are saved in the browser. The eth0 detail includes mirrored throughput, optics ranges, example errors and driver counters, plus counters, optics, addresses, queues and configuration tabs. These are static illustrative snapshots, not live router readings.

**System → Resources** and **Dataplane (ASK2) → Queues & pools** have range-aware heatmaps. Select a cell to open its sample and copy the URL to revisit it. **Logs & Events** shows entry counts as a histogram above the simulated journal replay; changing range updates the histogram, not the captured replay.

Traffic, firewall, conntrack, routing, WireGuard, QoS, ASK2, and system views include additional example snapshot summaries and related evidence links. Configuration diff and journal views link to example history and events. These panels are supplied by the provider, and disappear in the unavailable scenario. The traffic chart is an aggregate example, not the stacked per-client chart or Sankey in the mockup.

The firewall rules and zone views show an illustrative directional zone-policy matrix. Select a colored cell to filter its example default policy and placeholder exception rows; counts do not represent a router's rules or counters.

### Replacing Demo Data

The HTTP handler accepts a read-only [`Provider`](internal/demo/pages.go) via `HandlerWithProvider`. The default [`Fixtures`](internal/demo/fixtures.go) implementation returns in-memory snapshots; the HTTP layer owns navigation, filtering, sorting, and presentation. A router-backed adapter can implement the same contract, honor the selected historical range, return `ErrNotFound` for absent entities, and return no records or samples for unavailable sources. Sanitize or redact sensitive data before constructing snapshots. The current shell still identifies itself as a local demo; do not serve real snapshots from it until the production authentication and deployment boundaries are in place.

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

`mise.toml` pins Go and Node; npm and Go dependencies are locked in `package-lock.json`, `go.mod` and `go.sum`. The environment variables keep mise's downloaded toolchains, state and cache in this project. Set them in each shell before running `mise`; Go's normal module cache and npm's cache may remain in their standard user locations. No router-side Node installation is needed.

To run the browser and accessibility smoke test, install the pinned Playwright Chromium into the project and run it:

```sh
export PLAYWRIGHT_BROWSERS_PATH="$PWD/.tools/playwright"
mise exec -- npx playwright install chromium
mise exec -- make browser
```

On Linux, Playwright may require OS libraries; consult its installation documentation. The browser suite starts a loopback demo server, tests desktop and mobile layouts, and runs axe. `make check` builds the demo's UnoCSS utilities and also tests static Go cross-compilation for Linux arm64 and amd64 using a temporary fixture. The generated demo CSS is embedded in the binary; rebuild it after changing template utilities. Other build outputs live in `dist/`.

`mise exec -- make security` runs the pinned `govulncheck` on project code and the Go SBOM tool, plus `npm audit` (both query online vulnerability databases); `mise exec -- make sbom` writes separate Go and npm CycloneDX inventories under `dist/`. The Go scan concerns reachable code, not a vulnerability assessment of all installed tools. esbuild, svg-sprite, Font Awesome, uPlot and ECharts are available locally for the eventual asset pipeline. Production asset-size, release and packaging gates remain to be added; no `.deb` or production release exists yet.

The brief's template engine (`templ` or `html/template`), Datastar client/SDK version and CSP compatibility, journal strategy, ASK2 contract, icon subsetting, asset hashing/compression and production bundling remain decisions or future implementation work. The npm Datastar package's latest published version is a beta; do not assume it matches the Go SDK without validating the selected releases.

## Documentation

- [docs/TECH-STACK.md](docs/TECH-STACK.md) — Complete technology list required on the router (gather, store, Datastar presentation), OS/kernel prerequisites, `CGO_ENABLED=0` trade-offs, and CI-only build toolchain.
- [docs/STORAGE-FORMAT.md](docs/STORAGE-FORMAT.md) — Binary layout and engineering specification for the single-file multi-ring datastore (`vyview.ring`).
- [docs/DESIGN-BRIEF-ANALYSIS.md](docs/DESIGN-BRIEF-ANALYSIS.md) — Engineering analysis of the *VyOS View-Only Dashboard Design Brief v1.1 (FA Pro)*, collector cadence matrix, security checklist, licensing rules, and open architectural decisions.
- [plans/README.md](plans/README.md) — Design plans, architecture specifications, design system tokens, and dashboard screen mockups.
