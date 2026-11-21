# Architecture (§9; PDF pp. 47–52)

Static Go binary with embedded assets, collectors, state, rollups and web server; separate privileged read-only probe process. Official Datastar Go SDK, server-rendered HTML and one SSE stream per page. Template choice `html/template` versus `templ` remains open.

Scheduler uses ±10% jitter, one concurrent run per source and four concurrent execs globally, backoff to 5 min, on-demand subscriptions. Timeouts: 100 ms netlink, 2 s vtysh, 5 s op-mode. Parsers are JSON-first with release-specific text fallbacks and golden/fuzz tests. Normalize canonical units, join entities, redact before indexing; copy-on-write snapshots emit changes. RAM rings feed min/avg/max/p95 rollups, time-major store, event detector, search index and SSE hub.

## Collector catalogue

| Source | Preferred mechanism / fallback |
|---|---|
| Link/address/routes/neighbors | netlink (`200 ms` `IFLA_STATS64` + event-driven `rtnetlink` multicast) / `ip -j` |
| System & softnet | direct `/proc` (incl. `/proc/net/softnet_stat`), `/sys` (TMP431 `hwmon`, eMMC `life_time`) |
| nft, conntrack | read-only netlink through privileged helper (`CTNETLINK` `NEW`/`DESTROY` events + on-demand dump) / `nft -j`, `conntrack` |
| qdisc | RTM_GETQDISC/TCLASS / `tc -j -s` |
| WireGuard & IPsec | `wgctrl` (reconciled with ASK2 `wgfp_rx`; drop private keys at boundary); VICI + `NETLINK_XFRM` + `/proc/net/xfrm_stat` + CAAM stats |
| ethtool | netlink/ioctl (`ETHTOOL_GSTATS` incl. FMan MAC/RX FIFO overflows at 1–5 s; DOM every 300 s) |
| VRRP / keepalived | event-driven `keepalived` D-Bus signals (zero polling) |
| FRR | batched allowlisted `vtysh` JSON |
| VyOS op-mode | structured output or release-specific text; slow/on-demand |
| journal | one long-lived `journalctl -o json -f` (keeps `CGO_ENABLED=0`) |
| Kea | allowlisted read-only UNIX-socket JSON |
| Config/archive | on commit and inotify; read-only archive |
| DNS, chrony, LLDP, systemd | allowlisted exec/socket, D-Bus; slow cadences |
| ASK2 / DPAA1 & Two-Plane Reconciler | `fe_flow` debugfs, `fsl_dpa`/`qbman`, CAAM, `wgfp_rx` via `vyview-probe` (defensive schema guards) joined against `netdev`/`conntrack`; optional VPP stats segment |

## Execution and API

Only compiled allowlisted absolute argv templates, no shell; typed placeholders must refer to known entities. Interface names match `^[a-z][a-z0-9.\-]{0,14}$`; revisions are bounded archive integers; neighbors are configured addresses. Per-command timeout, output cap (example 16 MB), parser and minimal `LANG=C`/fixed PATH. Illustrative YAML sketch and paths below; verify each against its release.

```yaml
# /usr/share/vyview/allowlist.yaml (compiled into binary; review only)
- id: frr.bgp.summary
  argv: ["/usr/bin/vtysh", "-c", "show bgp vrf all summary json"]
  timeout: 2s
  max_output: 4MiB
  parser: frr.bgp_summary.v1
- id: ethtool.module
  argv: ["/usr/sbin/ethtool", "-m", "{iface:known_ethernet}"]
  timeout: 3s
  parser: ethtool.sff8472.text.v1
- id: opmode.dhcp.leases
  argv: ["/usr/libexec/vyos/op_mode/dhcp.py", "show_server_leases", "--family", "inet"]
  timeout: 6s
  parser: vyos.opmode.dhcp_leases.raw.v1
- id: config.show.json
  argv: ["/opt/vyatta/bin/vyatta-op-cmd-wrapper", "show", "configuration", "json"]
  timeout: 5s
  parser: vyos.config.json.v1
```

The op-mode function/flags and wrapper path depend on release.

Default on-box mode uses no VyOS API key. Optional remote mode limits compile-time endpoints to `/retrieve` (showConfig/returnValues/exists), allowlisted `/show`, and GET `/info`, with an outbound path filter and helper-only root-owned key. Keys are full-access, never read-only; forbidden mutation endpoints must be unreachable and tested.

## Datastar and SSE

Render complete initial HTML; morph changed fragments by stable ID; private `_chart` signals feed `<vy-chart>`/uPlot. Patch hashes avoid unchanged fragments; append logs, remove old rows, use SVG namespace as appropriate. Subscriptions key session, page, entities, tier and filter hash, sharing identical computation. Coalesce to cadence; bounded per-client queue (example 64 events/256 KB); overflow drops deltas and schedules snapshot; disconnect stalled clients after 30 s. Keepalive 15 s, no proxy buffering, no SSE compression; max 8 streams/session and 32 global. Retry 1 s, factor 2, max 30 s, full snapshot on every reconnect. Hidden-tab stream suspension is configurable. See Appendix B for illustrative snippets.

uPlot is primary; server SVG for sparklines/small heatmaps/topology; ECharts modular lazy-load for Sankey/large heatmaps. Build-time UnoCSS only; Font Awesome Free subset (~48 SVGs), attribution/NOTICE. Assets embedded, content-hashed, immutable and precompressed; HTML no-cache with ETag.

## Payload estimates

Initial brotli KB: HTML 9.5, Datastar 11, CSS 6, icons 8.5, uPlot 19.5, app JS 7, Inter 46, JetBrains Mono 24, manifest/icons 3 = 134.5 KB (64.5 with system fonts), against 150 KB budget. Lazy ECharts 118 KB and service worker 1.6 KB. These are estimates, not measured release sizes.
