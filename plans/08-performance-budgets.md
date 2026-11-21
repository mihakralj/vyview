# Performance budgets (§12; PDF pp. 60–62)

Planning estimates for LS1046A, not measurements: idle one-client CPU 0.99% of one core against <3%; RSS 30.1 MB against <40 MB; warm first paint <300 ms, cold <600 ms. Conntrack 50k dump while open adds estimated ~3% of one core. Never spawn processes at 1 s: estimated VyOS Python startup 650 ms.

| Source | Cost ms | Period s | Estimated % core |
|---|---:|---:|---:|
| `/proc` system | 0.12 | 1 | 0.012 |
| netlink links | 0.25 | 1 | 0.025 |
| softirq/interrupts | 0.3 | 5 | 0.006 |
| thermal | 0.15 | 5 | 0.003 |
| conntrack stats | 0.2 | 5 | 0.004 |
| nft counters | 2.5 | 5 | 0.050 |
| WireGuard | 0.3 | 5 | 0.006 |
| qdisc | 0.6 | 5 | 0.012 |
| ethtool stats | 0.8 | 5 | 0.016 |
| client accounting | 18 | 10 | 0.180 |
| FRR | 45 | 30 | 0.150 |
| Kea | 6 | 30 | 0.020 |
| DNS | 12 | 30 | 0.040 |
| chrony | 15 | 60 | 0.025 |
| op-mode | 650 | 300 | 0.217 |
| DOM | 40 | 300 | 0.013 |
| eMMC | 30 | 3600 | 0.001 |
| config | 25 | 3600 | 0.001 |
| journal | 0.05 | 1 | 0.005 |
| rollup | 25 | 60 | 0.042 |
| SSE rendering | 1.6 | 1 | 0.160 |

`GOMEMLIMIT=40MiB`, `GOMAXPROCS=2`, hard `MemoryMax=96M`; 50k-flow snapshot only while needed. SSE estimate 3,900 B/s; full render target <15 ms, fragments <0.5 ms.

Measure CPU <3% one client/<8% four, collector p95 within 2× estimates, RSS <40 MB and slope <1 MB/day over 72 h, flash host writes <50 MB/day over seven days, warm/cold paint, and no measurable throughput/drop impact under line-rate traffic. Detailed memory and measurement tables: [extraction](brief-extracted.txt), pp. 61–62.
