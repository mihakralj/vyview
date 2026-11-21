# Storage and retention (§10; PDF pp. 53–55)

Reference inventory: 692 series (158 core, 534 extended), including 12 interfaces, four cores, 50 clients, 120 firewall rules, 16 QoS classes, 12 WireGuard peers and optional ASK2. Full inventory is in [extraction](brief-extracted.txt), p. 53.

| Tier | Step | Span | Location | Series | Size |
|---|---|---|---|---:|---:|
| T0 | 1 s | 15 min | process RAM | 158 | 0.5 MB |
| T1 | 10 s | 48 h | tmpfs | 158 | 41.7 MB |
| T2 | 60 s | 14 d | flash | 692 | 212.9 MB |
| T3 | 300 s | 90 d | flash | 692 | 273.7 MB |
| T4 | 1 h | 2 y | flash | 692 | 185.0 MB |

T0 raw float32; other slots min/avg/max/p95 (16 bytes), rates for counters, gaps via validity/NaN. Full flash profile 671.6 MB, lean 412.4 MB. Cap top-50 clients plus other, 120 rules plus pinned, 32 peers and selected ethtool counters; show cap exhaustion. Query ≤1,500 points per series.

Recommended fixed-size time-major RRD-style blocks with CRC, batched 60 s, versus SQLite WAL or bbolt; remote storage optional. Estimated host writes 31.8 MB/day, NAND 95.4 MB/day at 3× amplification; SQLite estimate 1,108.7 MB/day host and naive series-major 3,892.5. Figures are estimates and require hardware measurement. `fsync` once per flush (≤60 s loss), event batches at most once per 5 s. T1 tmpfs is lost on reboot; optional clean-shutdown snapshot.

Recommended reference persistence: dedicated data partition/LV mounted across image upgrades. `/config/vyview` survives upgrades but is small, enters backups and is recommended only for lean profile if used. Image-local `/var/lib` loses history on upgrade; tmpfs-only loses history on reboot. Authenticated `/metrics`, exporter scraping and remote-write are optional; standalone operation remains required.
