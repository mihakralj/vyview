# Overview (§1–3; PDF pp. 5–10)

vyview is a single-router, LAN-only, read-only, real-time and historical VyOS dashboard. It runs on the router as a static Go application with Datastar, UnoCSS, Font Awesome Free and uPlot. Reference target: NXP LS1046A, four Cortex-A72 cores, arm64, Linux 6.18 and VyOS 1.5/rolling. Optional platform modules declare capabilities and degrade gracefully.

## Goals and boundaries

Answer common questions in at most three clicks without SSH; show history and changes with evidence; expose op-mode and redacted configuration; run on embedded ARM64 and x86 without external dependencies; observe optional ASK2 offload; support desktop, phone and wall displays.

No configuration writes, operational actions, shell or arbitrary commands, user-triggered diagnostics in v1, outbound notifications, fleet management, internet exposure, full DPI/Suricata management, or replacement of external long-term analytics. Future diagnostics, notifications and federation are separately risk-gated.

## Personas and jobs

Owner/architect validates commits, kernel and offload changes; troubleshooting power user follows symptoms to evidence; family IT admin glances at status and top talkers; kiosk user needs long-lived rotation; developer needs fake data, golden tests and measurable budgets.

Jobs include slow internet, bandwidth users, VPN state, impact of a commit, hardware offload, IoT reachability, BGP state, router health and flash wear. The brief maps each to specific navigation and evidence on pp. 7–8.

## Success metrics

| Metric | Target and measurement |
|---|---|
| Time to answer | ≤3 clicks or one search; Playwright journeys and three-user usability test |
| First paint | Warm <300 ms, cold <600 ms; LAN Playwright |
| Idle CPU | <3% of one A72 core, one Overview client, 10-minute pidstat |
| RSS | <40 MB after 72-hour soak |
| Flash | Batched rollups only; <50 MB/day host writes |
| Update latency | <1.5 s, synthetic collector to browser |
| Accessibility | WCAG 2.2 AA; zero serious axe violations |
| Stability | 72-hour SSE soak, four clients, reconnect storms; RSS slope <1 MB/day |

## Binding principles

Read-only by construction; source chips for every widget; overview, filter and drill-down with retained context; progressive disclosure; one shared page range, synchronized crosshairs, brushing, compare and annotations; live with pause and pending updates; five statuses with text and shapes as well as color; comfortable/compact/wall densities; WCAG 2.2 AA and keyboard operation; self-contained offline LAN operation. See the [extracted text](brief-extracted.txt) pp. 9–10 for the specific rules and accessibility thresholds.
