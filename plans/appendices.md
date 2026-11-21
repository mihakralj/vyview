# Appendices (PDF pp. 68–76)

The complete source text is preserved in [brief-extracted.txt](brief-extracted.txt), including all rows and code sketches. Page markers correspond to the PDF; broken line wraps are extraction artifacts.

## A — Data-source catalogue (pp. 68–69)

Maps each metric to reference command/file, mechanism/parser, cadence, retention and consuming pages. It covers interfaces, addresses, ethtool, DOM, neighbors, routes, BGP/OSPF/BFD, firewall, NAT, conntrack, clients, offload, qdisc, VPN, DHCP/DNS/NTP, LLDP, SSH, WAN/DDNS, CPU/memory/sensors/storage/eMMC, IRQs, processes, units, version, images, config, logs, probes and ASK2/Suricata. ASK2 entries are explicitly illustrative/TBD. The architecture and page plans provide focused summaries; the original catalogue's full table is in the extraction.

## B — Datastar snippets (pp. 70–71)

Illustrative only: page signals and `data-init`, retry settings, range/live controls, HTML fragment morphs, private chart signals, SSE events, Go handler and uPlot custom element. Verify against the pinned SDK version. Exact sketches are retained in the extraction.

## C — UnoCSS configuration sketch (p. 72)

Illustrative build-time `uno.config.ts`: filesystem template/JS scanning, `presetWind3`, optional FA Iconify collection, variant group transformer, CSS-variable theme colors, font families, card/chip/badge/table/KPI shortcuts and safelisted dynamic classes. The extraction preserves the complete sketch.

## D — Font Awesome icon map (p. 73)

Includes section/navigation/status/action icons and sprite IDs, including regular clock and copy. See the complete table in the extraction. Use Font Awesome Free and preserve required notices.

## E — URL and route map (p. 74)

GET routes: `/`, `/health`, `/interfaces[/{name}]`, `/clients[/{id}]`, `/conntrack`, `/routing` and subviews, `/firewall` and subviews, `/nat`, `/vpn` subviews, `/services/{service}`, `/qos[/{policy}]`, optional `/dataplane` subviews, `/system` subviews, `/config` and history, `/logs`, `/events`, `/diagnostics`, `/topology`, `/settings`, `/search`, `/export/{widget}`, `/sse/{page}`, `/a/{hash}.{ext}`, optional authenticated `/metrics`, and `/healthz`. `/login` and `/logout` are GET/POST authentication exceptions outside the data plane. Exact parameters and routes are in the extraction.

## F — Glossary (p. 75)

Defines AQM, ASK2, BFD, BMan/QMan/FMan, bufferbloat, CAKE, CEETM, conntrack, DOM, flowtable, FIB/RIB, morphing, op-mode, p95, PSI, rollup, signals, SSE, tier, time-major layout, VRF and flash write amplification. Full definitions are in the extraction.

## G — References (p. 76)

Datastar: https://data-star.dev/docs and https://github.com/starfederation/datastar; VyOS API: https://docs.vyos.io/en/rolling/automation/vyos-api.html; VyOS HTTPS: https://docs.vyos.io/en/rolling/configuration/service/https.html; UnoCSS: https://unocss.dev; Font Awesome: https://fontawesome.com; uPlot: https://github.com/leeoniya/uPlot; ECharts: https://echarts.apache.org; WCAG 2.2: https://www.w3.org/TR/WCAG22/; Shneiderman's visual information-seeking mantra; RFC 5737 and RFC 3849.
