# vyview Single-File Multi-Ring Datastore Specification (`vyview.ring`)

> Refines Design Brief §10 ("Time-series storage and retention") to enforce the **two-file router footprint**: **one static Go binary** (`/usr/bin/vyview`) and **one pre-allocated persistent data file** (`/var/lib/vyview/vyview.ring`), while supporting fluid **200 ms (5 Hz)** live telemetry with strict integer tier cascades.

---

## 1. Design Goals & Constraints

1. **Single persistent file on disk:** All persistent time-series tiers (T2/T3/T4), series metadata/catalog, curated event/alert/probe history, and an optional clean-shutdown snapshot of the hot tier live inside **one fixed-size binary file** (`vyview.ring`).
2. **Fluid real-time telemetry decoupled from flash I/O:** Fast kernel counters (interface RX/TX bytes & packets, CPU/softirq, active queue drops) are sampled at **200 ms (5 Hz)** into an in-process RAM ring (`T0`) and cascaded cleanly with exact integer multipliers (`25× → 12× → 5× → 12×`) into `T1` (`5 s`, RAM/`tmpfs`) and the persistent flash tiers (`T2`/`T3`/`T4`). Slower metrics (`1 s`+) piggyback on every 5th tick (`5 × 200 ms = 1 s`).
3. **Zero growth, zero fragmentation, zero compaction:** Pre-allocated once via `fallocate(2)` at initialization. File size never changes after creation.
4. **Minimal eMMC/SD flash wear (`< 50 MB/day` host writes):**
   - Time-major layout on disk (all series for a single timestamp slot are contiguous on disk).
   - Every disk slot is **4 KiB page-aligned** (matching Linux page size & eMMC flash page boundaries).
   - Batched writes once per 60 seconds using `pwrite(2)` + a single `fdatasync(2)` (no file-size change → minimal ext4 metadata/journal overhead).
   - Self-describing slots `(timestamp, slot_seq, valid_bitset, crc32c)` eliminate per-minute superblock rewrites.
5. **Crash-safe and power-cut resilient:** Per-slot CRC32C detects torn writes; ping-pong double-buffered superblock protects structural metadata; worst-case data loss on sudden power cut is bounded to `≤ 60 s` (or `≤ 5 min` in compact profile).
6. **Bounded RSS (`< 40 MB` total process RSS):** Uses explicit `pread(2)` / `pwrite(2)` with pooled buffers rather than `mmap(2)`, preventing unbounded kernel page-cache pressure and eliminating `SIGBUS` crash risk on I/O errors.
7. **Zero-migration series additions up to capacity:** Fixed-capacity series catalog with spare headroom (`760` series capacity vs `692` reference inventory) lets new interfaces, WireGuard peers, or pinned firewall rules claim free slots without rewriting the file.

---

## 2. Memory Tiers vs. Single Persistent File

Every tier is an **exact integer multiple** of the tier below it (`200 ms × 25 = 5 s`, `5 s × 12 = 60 s`, `60 s × 5 = 300 s`, `300 s × 12 = 3600 s`), allowing lightweight running accumulators with zero timestamp interpolation or resampling drift, and clean UTC midnight alignment across all persistent buckets.

| Tier | Step ($\Delta t$) | Fan-In | Slots ($N$) | Span / UI Windows Served | Series Class | Storage Location | Size |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **T0 Live Ring** | `200 ms` (5 Hz) | Base tick | `2,048` | `6 min 50 s` (live 2 m / 5 m 60 fps scroll) | Core (`≤ 192`, scalar `float32`) | In-process RAM (contiguous 4 KiB-aligned slab) | `~1.5 MiB` |
| **T1 Hot Ring** | `5 s` | `25 × T0` | `17,280` | `24 h` (15 m / 1 h / 6 h / 24 h high-res) | Core (`≤ 192`, `min/avg/max/p95`) | RAM / `tmpfs` (`/run/vyview/hot.ring`, zero flash wear) | `~50 MiB` *(in `tmpfs`)* |
| **T2 Warm Ring** | `60 s` | `12 × T1` | `20,160` | `14 d` (1 h / 6 h / 24 h / 7 d / 14 d) | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 3)** | `236.25 MiB` (full) / `157.50 MiB` (lean) |
| **T3 Cool Ring** | `300 s` | `5 × T2` | `25,920` | `90 d` (30 d / 90 d) | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 4)** | `303.75 MiB` (full) / `202.50 MiB` (lean) |
| **T4 Cold Ring** | `3600 s` | `12 × T3` | `17,520` | `2 y` (1 y / 2 y + hour-of-week anomaly bands) | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 5)** | `205.31 MiB` (full) / `136.88 MiB` (lean) |
| **Event / Alert / Probe Ring** | Event | — | `4,096` frames | `30–90 d` events, alerts, probes, `mtr` paths | Structured records | **`vyview.ring` (Region 6)** | `16.00 MiB` |
| **T1 Shutdown Snapshot** *(opt.)* | On stop | — | 1 zstd blob | Latest `24 h` hot tier across planned reboots | Snapshot of T1 on `SIGTERM` | **`vyview.ring` (Region 7)** | `16.00 MiB` |

> **Storage Profiles (Single-File Size on Disk):**
> - **Full profile (`profile = 1`):** `min/avg/max/p95` for all 760 series across T2/T3/T4 — **`~778 MiB`** (`745.3 MiB` TS rings + `32.3 MiB` headers/catalog/events/T1-snapshot).
> - **Lean profile (`profile = 2`):** `min/avg/max/p95` for 192 core series; `avg/max` for 568 extended series — **`~529 MiB`** (`496.9 MiB` TS rings + `32.3 MiB` headers/catalog/events/T1-snapshot).
> - **Compact profile (`profile = 3`, small-flash / SD-card friendly):** Keeps `T0` (`200 ms × 2,048`) and `T1` (`5 s × 2,048` = `2 h 51 m`, `~6 MiB` RAM) in memory, omits the 60 s disk tier, and persists only `T2 = 5 min × 2,048` (`7.11 d`, `24 MiB` full / `16 MiB` lean) and `T3 = 1 h × 8,784` (`366 d` leap year, `103 MiB` full / `69 MiB` lean) — **`~160 MiB` total** (`~117 MiB` lean) and **`< 5 MB/day`** host writes.
> - **Live-only profile:** No persistent flash file; `T0` + `T1` run purely in RAM/`tmpfs` (zero flash writes).

### 2.1 In-RAM `T0` Slab Layout & Fast/Slow Loop Split

- **Contiguous page-aligned `T0` backing slab:** `T0` uses `N = 2,048` `float32` samples per core series (`2,048 × 4 B = 8,192 B` = exact `2 × 4 KiB` pages per series, and power-of-two index masking `idx & 2047`). Allocated as one contiguous backing slice (`192 × 8 KiB = 1.5 MiB`) to eliminate heap fragmentation and minimize TLB pressure on Cortex-A72.
- **Fast loop (`200 ms` / 5 Hz):** Reads interface `IFLA_STATS64` via netlink `RTM_GETLINK`, `/proc/stat` + `/proc/softirqs`, and active queue drop counters. Most NIC drivers and NXP DPAA1/FMan hardware offload engines coalesce shadow counter updates on ~100–250 ms cycles, so 200 ms avoids artificial `0`-byte gaps while eliminating the jerky 1-second visual step.
- **Slow loop (`1 s`+ piggyback):** Thermals, fan RPM, memory, conntrack counts, and control-plane state piggyback on every 5th tick (`1 s`), 25th tick (`5 s`), or 150th tick (`30 s`), writing directly into `T1` or `T2` accumulators without inflating `T0`.
- **Optional 100 ms (10 Hz) diagnostics burst mode:** When explicitly enabled on a local LAN interface detail / diagnostics view (e.g., watching `iperf3` or `TRex` bursts), the UI stream may temporarily run at 100 ms in-memory without altering the persistent `T1`–`T4` ring geometry.

---

## 3. Binary File Layout (`vyview.ring`)

All integers are **little-endian**. Every region starts on a **4 KiB (`4096 B`) boundary**.

```
Offset (4 KiB aligned)
┌─────────────────────────────────────────────────────────────────────────┐
│ Region 0: Superblock A (4 KiB) + Superblock B (4 KiB)      [8 KiB]      │
│   Magic "VYVW", format version, profile, generation++, CRC32C,          │
│   region table (offsets, slot sizes, capacities), T1 snapshot metadata  │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 1: Series Catalog A (128 KiB) + Backup B (128 KiB)  [256 KiB]    │
│   1,024 fixed 128-byte descriptors: series_id, canonical key hash+name, │
│   unit, metric kind (gauge/counter), tier class, active/retired flags   │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 2: Reserved / Padding to 1 MiB boundary             [760 KiB]    │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 3: Ring T2 — Warm  (60 s × 20,160 slots = 14 d)     [236.25 MiB] │
│   Contiguous circular array of page-aligned time-major slot blocks      │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 4: Ring T3 — Cool  (300 s × 25,920 slots = 90 d)    [303.75 MiB] │
│   Contiguous circular array of page-aligned time-major slot blocks      │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 5: Ring T4 — Cold  (3600 s × 17,520 slots = 2 y)    [205.31 MiB] │
│   Contiguous circular array of page-aligned time-major slot blocks      │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 6: Circular Event / Alert / Probe Ring              [16.00 MiB]  │
│   4 KiB framed blocks of length-prefixed records (events, alert state   │
│   transitions, traceroute paths, latency-under-load / speed-test runs)  │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 7: Optional T1 Hot-Ring Clean-Shutdown Snapshot     [16.00 MiB]  │
│   Zstd-compressed T1 ring written only on SIGTERM / ExecStop; validated │
│   by CRC32C and freshness timestamp on next startup                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### 3.1 Superblock Structure (4 KiB, Ping-Pong A/B)

Written **only** on file creation, series catalog mutation, or clean shutdown (never on routine 60 s metric flushes):

| Field | Size | Description |
| --- | --- | --- |
| `magic` | `4 B` | ASCII `"VYVW"` |
| `format_version` | `4 B` | Schema version (starts at `1`) |
| `generation` | `8 B` | Monotonically increasing counter; higher valid CRC wins on mount |
| `profile` | `4 B` | `1 = full`, `2 = lean`, `3 = compact` |
| `max_series` | `4 B` | `760` (fits exact 4 KiB slot geometry with zero wasted padding) |
| `core_series_cap` | `4 B` | `192` |
| `flags` | `4 B` | Bit 0: `CLEAN_SHUTDOWN`, Bit 1: `T1_SNAPSHOT_VALID` |
| `created_unix_ns` | `8 B` | File creation timestamp |
| `last_shutdown_unix_ns` | `8 B` | Timestamp of last clean shutdown |
| `regions[8]` | `8 × 48 B` | Per-region descriptor: `kind`, `offset`, `length`, `slot_bytes`, `slot_count`, `step_sec` |
| `catalog_crc32c` | `4 B` | CRC32C of active Series Catalog region |
| `t1_snapshot_len` / `crc` | `12 B` | Compressed length + CRC32C of Region 7 snapshot |
| `reserved` | padded | Zero-filled to byte `4092` |
| `header_crc32c` | `4 B` | Castagnoli CRC32C over bytes `[0..4092)` |

### 3.2 Time-Major Slot Block Geometry (Zero Wasted Padding)

Every time slot in T2, T3, and T4 is an atomic, page-aligned, self-verifying block:

#### Full Profile Slot (`12,288 B` = exact `3 × 4 KiB` pages)
- **Slot Header (`128 B`):**
  - `slot_magic` (`2 B` = `0x5653`), `flags` (`2 B`), `crc32c` (`4 B`, over entire slot excluding this field)
  - `unix_sec` (`8 B`, aligned bucket start timestamp, e.g., `ts - ts%step`)
  - `slot_seq` (`8 B`, strictly monotonic write sequence counter per ring)
  - `valid_bits` (`96 B` = 768 bits, 1 bit per series indicating valid sample vs gap/collection failure)
  - `reserved` (`8 B`)
- **Series Aggregates (`760 × 16 B = 12,160 B`):**
  - Per series: `[min, avg, max, p95]` as `4 × float32` (`16 B`).
  - `p95` computed in-memory during the rollup interval via a fixed 64-bin log histogram or lightweight t-digest.
  - **Monotonic $\Delta t$ rate calculation:** At 200 ms sampling, even 5–10 ms of OS scheduler jitter creates a 2.5–5% rate error if dividing by a hardcoded interval. Every counter read (interface bytes/packets, firewall hits, QoS drops) MUST be stamped with Go's monotonic clock (`time.Now()`) and converted to per-second rates using the exact elapsed monotonic nanoseconds:
    $$\text{Rate} = \frac{\Delta \text{counter}}{\Delta t_{\text{monotonic, sec}}}$$
    with 32/64-bit counter wrap, interface-reset, and ASK2 debugfs driver-reload counter-reset detection applied before aggregation.
  - **Gap-safe running accumulators (`validCount`):** Each active rollup bucket maintains `{validCount, sum, min, max, histogram}` in RAM. When a bucket closes, `avg = sum / validCount`. If `validCount == 0` (e.g., interface down, collector timeout, or reboot downtime), the series' bit in `valid_bits` is cleared and `math.NaN()` is stored so charts render honest gaps instead of dragging averages down with synthetic zeros.
  - **Peak preservation (`max` alongside `avg`):** Because averaging `1,500+` sub-second samples into multi-minute buckets flattens short multi-gigabit bursts, every tier from `T1` upward stores `max` alongside `avg` so `uPlot` can plot the sustained average line with a translucent peak envelope behind it.
- **Total:** `128 + 12,160 = 12,288 B` (exact 3 pages).

#### Lean Profile Slot (`8,192 B` = exact `2 × 4 KiB` pages)
- **Slot Header:** `128 B` (same layout).
- **Core Series (`192 × 16 B = 3,072 B`):** `[min, avg, max, p95]` (`float32`).
- **Extended Series (`568 × 8 B = 4,544 B`):** `[avg, max]` (`float32`).
- **Padding:** `448 B` zero-filled to `8,192 B` (exact 2 pages).

### 3.3 Series Catalog Entry (`128 B` × `1,024` entries, Double-Buffered)

| Field | Size | Description |
| --- | --- | --- |
| `series_id` | `2 B` | Slot index `0..759` inside each time-major block |
| `class` | `1 B` | `0 = core`, `1 = extended` |
| `kind` | `1 B` | `0 = gauge`, `1 = counter_rate` |
| `unit` | `1 B` | Enum (`bps`, `pps`, `bytes`, `pct`, `celsius`, `ms`, `dbm`, `count`, `ratio`, ...) |
| `state` | `1 B` | `0 = free`, `1 = active`, `2 = retired` (retained for historical reads until overwritten) |
| `reserved` | `2 B` | Alignment |
| `key_hash` | `8 B` | XXH64 / FNV-64a of canonical key |
| `first_seen_sec` | `8 B` | First timestamp written |
| `last_seen_sec` | `8 B` | Last timestamp written (updated lazily in RAM, persisted on catalog change/shutdown) |
| `canonical_key` | `96 B` | UTF-8 key, e.g., `iface:eth0:rx_bps`, `client:mac:001132aabb01:rx_bps`, `rule:ipv4/fwd/120:pkts` |

Cardinality caps (Brief §10.2) enforce bounded series usage and memory under 8k–50k+ concurrent sessions:
- Top-50 clients by 24 h bytes + `other` bucket (`102` series), tracked via a bounded **Space-Saving / Heavy-Hitter Top-K sketch** so RAM stays constant regardless of session cardinality
- Top-120 firewall rules by hits + pinned rules (`240` series)
- Up to 32 WireGuard peers (including `wgfp_rx` fast-path counters when ASK2 is active)
- Two-plane offload ratios (`offload_ratio` per port, plus aggregate HW/SW flow counts and CAAM job rates)
- `ethtool -S` history only for pinned or well-known MAC/FIFO error counters (`≤ 32` per interface)

### 3.4 Circular Event / Alert / Probe Ring (Region 6, `16 MiB`)

Replaces separate append-only event log files and probe caches:
- Divided into `4,096` page-aligned `4 KiB` frames (`frame_magic`, `frame_seq`, `unix_sec`, `used_bytes`, `crc32c`, followed by packed binary/CBOR records).
- Record types:
  1. **Curated Events** (commits, reboots, link flaps, BGP/OSPF/BFD transitions, WG stale, DHCP pool alerts, systemd unit failures) — fuels chart annotations and 30 d Event Timeline.
  2. **Alert Transitions** (raised/cleared, check ID, entity, value vs threshold) — 90 d Alert History.
  3. **Diagnostic Probe Results** (scheduled `mtr` hop paths for 30 d path diffs, latency-under-load bufferbloat grades & nightly speed tests for 90 d).
- Flushed at most once every 5 s when dirty, or batched alongside the 60 s rollup `fdatasync`.

---

## 4. Write Path, Recovery & Flash-Wear Math

### 4.1 Normal Write Path (Every 60 s)
1. Collectors sample the fast loop every `200 ms` into `T0` (in-process RAM) using monotonic $\Delta t$, rolling every `25` ticks into `T1` (`5 s` `tmpfs` ring) with zero flash I/O. Slower collectors (`1 s`+) update `T1`/`T2` accumulators directly.
2. Every `60 s` (`12 × T1`), the Rollup Engine builds:
   - 1 × T2 slot (`12 KiB` full / `8 KiB` lean) at index `(unix_sec / 60) % 20160`
   - Every `5 min` (`5 × T2`): 1 × T3 slot at `(unix_sec / 300) % 25920`
   - Every `60 min` (`12 × T3`): 1 × T4 slot at `(unix_sec / 3600) % 17520`
3. Write slots via `pwrite(2)` and call **`fdatasync(fd)` once**.
4. Because the file size never changes and the superblock is untouched during normal operation, ext4 only commits the written data blocks and minimal journal metadata.

### 4.2 Daily Flash-Wear Estimate (Full Profile, 760-Series Capacity)

| Operation | Frequency / Day | Payload / Write | Payload / Day |
| --- | --- | --- | --- |
| T2 Warm ring (`60 s`) | `1,440` | `12 KiB` | `16.88 MiB` |
| T3 Cool ring (`300 s`) | `288` | `12 KiB` | `3.38 MiB` |
| T4 Cold ring (`3600 s`) | `24` | `12 KiB` | `0.28 MiB` |
| Event/alert/probe frames | `~200` (typical) | `4 KiB` | `0.78 MiB` |
| **Total payload** | | | **`~21.3 MiB/day`** |
| **Est. host writes (incl. ext4 journal)** | | | **`~28–32 MB/day`** (budget: `< 50 MB/day`) |
| **Est. NAND writes (eMMC WAF 3×)** | | | **`~85–96 MB/day` (`~34 GB/year`)** |
| **10-year wear on 16 GB MLC eMMC (~46.9 TBW)** | | | **`~0.71%` of endurance** |

*(In the **lean profile** with `8 KiB` slots, payload drops by one-third to `~14.5 MiB/day`. In the **compact profile** omitting the 60 s disk tier, payload drops to **`~4.4 MiB/day`**.)*

### 4.3 Startup & Crash Recovery (Zero Replay Needed)
1. Read Superblock A and B; select the one with valid `header_crc32c` and highest `generation`.
2. Load active Series Catalog (`128 KiB`, verified by `catalog_crc32c`).
3. For each ring (T2, T3, T4, Event Ring), locate the newest valid slot head via **O(log N) binary search** on `(slot_seq, crc32c)` (~15 `pread`s per ring, `< 2 ms` total on eMMC). Any single torn slot at the head fails CRC and is cleanly skipped/overwritten on the next tick.
4. If `CLEAN_SHUTDOWN` and `T1_SNAPSHOT_VALID` are set and snapshot age `< 24 h`, decompress Region 7 into `/run/vyview/hot.ring` to restore 5 s resolution across planned reboots.

---

## 5. Query Path & Downsampling

1. **Tier selection** based on requested window `[from, to]` (Brief §7.5):
   - `≤ 6 min` (Live 2 m / 5 m) → **T0** (`200 ms` in-process RAM; streamed at 5 Hz and rendered at 60 fps in `<vy-chart>`)
   - `≤ 24 h` → **T1** (`5 s` `tmpfs` RAM; falls back seamlessly to T2 across reboot gaps or in compact profile beyond `2 h 51 m`)
   - `≤ 14 d` → **T2** (`60 s` in `vyview.ring`; `300 s` in compact profile up to `7 d`)
   - `≤ 90 d` → **T3** (`300 s` in `vyview.ring`)
   - `> 90 d` (up to `2 y`) → **T4** (`3600 s` in `vyview.ring`)
2. **Fast range reads:** Because slots are contiguous in time, reading a 24 h window from T2 (`1,440` slots) is a single sequential scan of `~17 MiB` (or wrap-around split into 2 sequential reads) that extracts all requested series simultaneously in one pass.
3. **Max-min preserving downsampling (M4 / LTTB):** Server caps output to `~1,000–1,500` points per series (matching typical 600–1,200 px chart width ~1:1) so `uPlot` renders in `< 1 ms` with zero client-side decimation and tiny SSE payloads.
4. **Anomaly bands (Brief §7.13):** Computed server-side over the same hour-of-week buckets in T3/T4 using robust median + MAD (`baseline ± kσ`).

---

## 6. Upgrades, Persistence & Self-Monitoring

- **Recommended path on VyOS:** `/var/lib/vyview/vyview.ring` mounted on a dedicated eMMC partition or LV (survives both reboots and VyOS image upgrades without bloating `/config` backups).
- **Schema upgrades:** On a `format_version` bump, `vyview` runs an offline/startup in-place or streaming migration tool (or starts a fresh `vyview.v2.ring` while mounting the prior file read-only for seamless historical queries until it ages out).
- **Self-measurement:** `vyview` reads `/proc/self/io` (`write_bytes`) and exposes real write rate, ring fill/oldest timestamp, cardinality cap saturation, and eMMC `EXT_CSD` life-time estimates on `System › Storage` and `Settings › Retention`.
