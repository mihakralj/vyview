# Dashboard features and mockups (§7–8; PDF pp. 28–46)

Server owns data and HTML; client owns ephemeral state. Features: searchable keyboard palette and global index; synchronized chart cursors and brushing; range presets 5 m, 1 h, 24 h, 7 d, 30 d, 90 d and absolute; previous-period comparison; annotations; pause and pending counts; server SVG sparklines and small heatmaps; lazy ECharts for Sankey and large heatmaps; uPlot stacked series and anomaly bands; metric explanations, copy-as-CLI, CSV/JSON/PNG export, deep links, saved views, keyboard shortcuts, kiosk, responsive/mobile, PWA shell, themes/density, skeletons, stale/connection indicators, diff and SVG topology. The complete behavior/implementation table is in [extraction](brief-extracted.txt), pp. 28–29.

Range mapping: ≤15 min live ring; ≤48 h 10 s; ≤14 d 1 min; ≤90 d 5 min; otherwise 1 h. Cap ~1,500 points/series with min/max-preserving downsampling.

## Mockups

All data and ASK2 sources are illustrative; public IPs use documentation ranges. Each link opens a PNG rendering of the corresponding full PDF page, including its numbered callouts. The source pages are also available in the [design brief](VyOS-ViewOnly-Dashboard-Design-Brief.pdf), pp. 31–46, and the callout text is in the [extraction](brief-extracted.txt).

| Mockup | Callouts |
|---|---|
| [M1 Overview](mockups/m01-overview.png) | Conditional alerts; WAN evidence; event markers and anomaly band; KPI status; events |
| [M2 Interfaces](mockups/m02-interfaces.png) | Multi-select chips; columns; one-hour sparks; error highlights; source footer |
| [M3 Interface detail](mockups/m03-interface-detail.png) | Tabs; mirrored throughput and flap; errors; DOM thresholds; driver counters |
| [M4 Traffic](mockups/m04-traffic.png) | Top six + other; Sankey; identity sources; randomized MAC privacy |
| [M5 Firewall](mockups/m05-firewall.png) | Chains; rule sparks and shorthand; zone matrix; rule drawer |
| [M6 Conntrack/NAT](mockups/m06-conntrack-nat.png) | Offload totals; validated filters; virtual rows; HW/SW path; flow drawer |
| [M7 Routing](mockups/m07-routing.png) | Neighbor/BFD state; diagnosis; longest match; RIB churn |
| [M8 WireGuard](mockups/m08-wireguard.png) | Optional acceleration; stale thresholds; public keys only; age chart |
| [M9 QoS](mockups/m09-qos.png) | Bufferbloat grade; CAKE tins; probe comparison; backlog |
| [M10 ASK2](mockups/m10-ask2.png) | Capability notice; HW/SW flow; CEETM heatmap; BMan pool; Suricata shunt |
| [M11 System](mockups/m11-system.png) | CPU heatmap; eMMC wear; image and boot history |
| [M12 Configuration](mockups/m12-configuration.png) | Tree; diff modes; masked changes; offline explain |
| [M13 Logs](mockups/m13-logs.png) | Histogram; paused tail; pending count; cursor permalinks |
| [M14 Palette](mockups/m14-palette.png) | Cross-entity query; client hit; state/config evidence; keyboard hints |
| [M15 Mobile](mockups/m15-mobile.png) | Alert banner; three-column KPIs; bottom nav; range; copyable detail |
| [M16 Topology](mockups/m16-topology.png) | Router/offload state; utilization bands; dashed tunnels; LLDP/BGP health |
