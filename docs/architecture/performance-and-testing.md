# Performance Budgets & Testing Strategy

> Defines vyview's quantitative resource budgets on the NXP LS1046A reference router, per-collector CPU estimates, and the automated testing and release gates required before shipping.

---

## 1. Resource & Latency Budgets (LS1046A Reference)

All figures below are target budgets for the **NXP LS1046A** (4× Cortex-A72, Linux 6.18, VyOS 1.5/rolling); planning estimates must be validated by hardware measurements in v0.1–v0.2.

| Resource / Metric | Hard Budget | Estimated / Notes | Measurement Method |
| --- | --- | --- | --- |
| **Idle-streaming CPU (1 Overview client)** | `< 3%` of 1× A72 core | `~0.99%` (`< 8%` with 4 clients) | 10-minute `pidstat` |
| **Conntrack 50k dump (while page open)** | `≤ +3%` of 1× A72 core | On-demand only (`5 s` while Conntrack tab open) | `pidstat` under 50k flows |
| **Resident memory (`VmRSS`)** | `< 40 MB` (`MemoryMax=96M`) | `~30.1 MB` (`GOMEMLIMIT=40MiB`, `GOMAXPROCS=2`) | 72-hour soak, slope `< 1 MB/day` |
| **Flash writes (host)** | `< 50 MB/day` | `~28–32 MB/day` full / `~19–22` lean / `< 5` compact | 7-day `/proc/self/io` + `iostat` |
| **Initial web payload (Brotli)** | `< 150 KB` | `~139.8 KB` (embedded IBM Plex) / `~69.8 KB` (`-tags systemfonts`) | CI asset-size gate |
| **First paint (LAN)** | Warm `< 300 ms`, cold `< 600 ms` | Full HTML render `< 15 ms`, fragment `< 0.5 ms` | Playwright LAN trace |
| **Counter-to-DOM latency** | `< 1.5 s` | `200 ms` fast loop + `250–300 ms` `60 fps` interpolation buffer | Timestamp delta probe |
| **SSE bandwidth (Overview)** | Low LAN overhead | `~5–10 KB/s` (`5 Hz` fast loop + `1 Hz` slow piggyback) | Browser DevTools / server stats |
| **Real-time / TSN forwarding no-harm** | `< X µs` p99 latency & jitter impact *(TBD by measurement)* | Zero measurable throughput drop or packet loss under line-rate load | Hardware traffic generator (`iperf3`/`TRex`) |

---

## 2. Per-Collector CPU Budget Breakdown (Planning Estimates)

Never spawn external processes on the fast (`200 ms` / `1 s`) loop — a single VyOS Python op-mode invocation costs `~650 ms` on Cortex-A72.

| Source | Mechanism | Cadence | Est. Cost / Run | Est. % of 1 Core |
| --- | --- | --- | ---: | ---: |
| `/proc` CPU, load, memory | Direct `/proc` read | `200 ms`–`5 s` | `0.12 ms` | `0.012%`–`0.060%` |
| Netlink links (`IFLA_STATS64`, 12 ifaces) | `RTM_GETLINK` | `200 ms` | `0.25 ms` | `0.125%` |
| IRQ / `softirqs` / `softnet_stat` | Direct `/proc` read | `5 s` | `0.30 ms` | `0.006%` |
| Thermal / `hwmon` (TMP431) | Direct `/sys` read | `5 s` | `0.15 ms` | `0.003%` |
| Conntrack stats | `CTNETLINK` stats | `5 s` | `0.20 ms` | `0.004%` |
| `nftables` counters (120 rules) | Netlink `NFT` dump | `5 s` | `2.50 ms` | `0.050%` |
| WireGuard peers (`wgctrl` + `wgfp_rx`) | Generic netlink | `5 s` | `0.30 ms` | `0.006%` |
| QoS qdisc / CAKE `xstats` | Netlink `RTM_GETQDISC` | `5 s` | `0.60 ms` | `0.012%` |
| NIC driver / FMan MAC stats (5 ports) | `ETHTOOL_GSTATS` | `5 s` | `0.80 ms` | `0.016%` |
| Per-client accounting (Top-K sketch) | `CTNETLINK` events + reconcile | `10 s` | `18.0 ms` | `0.180%` |
| FRR (BGP/OSPF/RIB/BFD) | Batched `vtysh` JSON | `30 s` (`5 s` BFD) | `45.0 ms` | `0.150%` |
| Kea DHCP leases & pools | UNIX control socket | `30 s` | `6.0 ms` | `0.020%` |
| PowerDNS Recursor | `rec_control get-all` | `30 s` | `12.0 ms` | `0.040%` |
| Chrony NTP | `chronyc -c` | `60 s` | `15.0 ms` | `0.025%` |
| VyOS op-mode (structured) | Allowlisted op-mode | `300 s` | `650.0 ms` | `0.217%` |
| SFP+ DOM optics (2 cages, I2C) | Ethtool module EEPROM | `300 s` | `40.0 ms` | `0.013%` |
| eMMC wear (`EXT_CSD`) | Kernel sysfs | `3600 s` | `30.0 ms` | `0.001%` |
| Running config & archive | `inotify` + gzip read | On commit | `25.0 ms` | `0.001%` |
| `journald` tail | `journalctl -o json -f` | Event-driven | `0.05 ms` | `0.005%` |
| Rollup flush + SSE rendering | In-process | `60 s` / live | `25.0` / `1.6 ms` | `0.042%` + `0.160%` |

*Gate:* Measured collector p95 duration on LS1046A must remain within `2×` of planning estimates.

---

## 3. Testing & Release Quality Gates

Every pull request and release candidate passes automated gates across parsers, security boundaries, accessibility, performance, and hardware parity:

| Gate Area | Approach & Pass Criterion | Cadence |
| --- | --- | --- |
| **Parser goldens** | Golden captures for every op-mode/JSON/text parser across VyOS `1.4 LTS`, `1.5`, and `rolling` on both `x86_64` and `LS1046A arm64`. | Every PR + weekly rolling capture |
| **Fuzzing & robustness** | Native Go fuzz tests (`go test -fuzz`) on all text parsers, netlink/RPC decoders, and Datastar signal decoders; zero panics. | `10 min/target` nightly |
| **Deterministic UI scenarios** | Fake collectors synthesize link flaps, BGP down, DHCP pool full, 50k conntrack flows, offload saturation, and unavailable sources. | Every PR |
| **Visual & responsive** | Playwright screenshots across all pages × dark/light themes × desktop/tablet/mobile breakpoints; pixel diff review. | Every PR |
| **Accessibility** | `axe-core`, keyboard-only navigation journeys, focus traps in drawers, contrast checks; **`0` serious or critical violations**. | Every PR |
| **Read-only enforcement** | Route-table test verifying `GET`/`HEAD` on data plane (accounting explicitly for `/login` and `/logout` POST exceptions); static check forbidding `os/exec` outside `internal/exec`; netlink message-type allowlist test. | Every PR |
| **Security & licensing** | CSP header checks, static Datastar `data-*` lint, secret-redaction tests covering all 19 secret classes, `govulncheck`, `npm audit`, and Font Awesome Pro leak guard on public builds. | Every PR |
| **MCP write-safety** | Automated suite verifying `100%` of MCP writes route through `commit-confirm`, `100%` appear in audit logs, and no arbitrary shell/command arguments are accepted. | Every PR (once MCP lands) |
| **Hardware-agnostic parity** | Full Playwright journey suite passes on both stock `x86_64` (no offload) and `LS1046A` (with ASK2). | Release gate |
| **Load & OT soak** | `50k` conntrack flows, `200` clients, `4` concurrent browsers within `2×` budgets; `72 h` reconnect-storm soak (`RSS` slope `< 1 MB/day`, no leaked goroutines/streams); MRP ring soak capturing `100%` of ring transitions in `< 2 s`. | Pre-release |
| **Performance & payload** | Go benchmarks for renderer, rollup, and search (`≤ 10%` regression); initial Brotli payload `< 150 KB`; palette search `< 50 ms` p95. | Every PR |
| **Upgrade & recovery** | Power-cut torn-slot recovery tests on `vyview.ring`; `.vN.ring` upgrade + read-only prior mount test; VyOS image upgrade persistence verification. | Pre-release |
