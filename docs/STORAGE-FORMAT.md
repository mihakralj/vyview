# vyview Single-File Multi-Ring Datastore Specification (`vyview.ring`)

> Refines Design Brief §10 ("Time-series storage and retention") to enforce the **two-file router footprint**: **one static Go binary** (`/usr/bin/vyview`) and **one pre-allocated persistent data file** (`/var/lib/vyview/vyview.ring`).

---

## 1. Design Goals & Constraints

1. **Single persistent file on disk:** All persistent time-series tiers (T2/T3/T4), series metadata/catalog, curated event/alert/probe history, and an optional clean-shutdown snapshot of the hot tier live inside **one fixed-size binary file** (`vyview.ring`).
2. **Zero growth, zero fragmentation, zero compaction:** Pre-allocated once via `fallocate(2)` at initialization. File size never changes after creation.
3. **Minimal eMMC/SD flash wear (`< 50 MB/day` host writes):**
   - Time-major layout (all series for a single timestamp slot are contiguous on disk).
   - Every slot is **4 KiB page-aligned** (matching Linux page size & eMMC flash page boundaries).
   - Batched writes once per 60 seconds using `pwrite(2)` + a single `fdatasync(2)` (no file-size change → minimal ext4 metadata/journal overhead).
   - Self-describing slots `(timestamp, slot_seq, valid_bitset, crc32c)` eliminate per-minute superblock rewrites.
4. **Crash-safe and power-cut resilient:** Per-slot CRC32C detects torn writes; ping-pong double-buffered superblock protects structural metadata; worst-case data loss on sudden power cut is bounded to `≤ 60 s`.
5. **Bounded RSS (`< 40 MB` total process RSS):** Uses explicit `pread(2)` / `pwrite(2)` with pooled buffers rather than `mmap(2)`, preventing unbounded kernel page-cache pressure and eliminating `SIGBUS` crash risk on I/O errors.
6. **Zero-migration series additions up to capacity:** Fixed-capacity series catalog with spare headroom (e.g., `760` series capacity vs `692` reference inventory) lets new interfaces, WireGuard peers, or pinned firewall rules claim free slots without rewriting the file.

---

## 2. Memory Tiers vs. Single Persistent File

| Tier | Step | Span | Slots | Series Class | Storage Location | Size |
| --- | --- | --- | --- | --- | --- | --- |
| **T0 Live Ring** | `1 s` | `15 min` | `900` | Core (`≤ 192`) | In-process RAM (`[]float32` lock-free ring) | `~0.6 MiB` |
| **T1 Hot Ring** | `10 s` | `48 h` | `17,280` | Core (`≤ 192`) | RAM / `tmpfs` (`/run/vyview/hot.ring`, zero flash wear) | `~42–50 MiB` |
| **T2 Warm Ring** | `60 s` | `14 d` | `20,160` | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 3)** | `236.25 MiB` (full) / `157.50 MiB` (lean) |
| **T3 Cool Ring** | `300 s` | `90 d` | `25,920` | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 4)** | `303.75 MiB` (full) / `202.50 MiB` (lean) |
| **T4 Cold Ring** | `3600 s` | `2 y` | `17,520` | Core + Extended (`≤ 760`) | **`vyview.ring` (Region 5)** | `205.31 MiB` (full) / `136.88 MiB` (lean) |
| **Event / Alert / Probe Ring** | Event | `30–90 d` | Variable records | Events, alerts, probes, mtr paths | **`vyview.ring` (Region 6)** | `16.00 MiB` |
| **T1 Shutdown Snapshot** *(opt.)* | On stop | Latest `48 h` | 1 zstd blob | Snapshot of T1 on `SIGTERM` | **`vyview.ring` (Region 7)** | `16.00 MiB` |

> **Total single-file size on disk:**
> - **Full profile** (`min/avg/max/p95` for all 760 series): **`~778 MiB`** (745.3 MiB TS rings + 32.3 MiB headers/catalog/events/T1-snapshot).
> - **Lean profile** (`min/avg/max/p95` for 192 core series; `avg/max` for 568 extended series): **`~529 MiB`** (496.9 MiB TS rings + 32.3 MiB headers/catalog/events/T1-snapshot).
> - **Live-only profile** (SD-card / zero-wear mode): No persistent flash file; T0 + T1 run purely in RAM/`tmpfs`.

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
| `profile` | `4 B` | `1 = full`, `2 = lean` |
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
  - Counters (interface bytes/packets, firewall hits, QoS drops) are converted to per-second rates with 32/64-bit wrap and interface-reset detection before aggregation.
  - Missing samples clear the series' bit in `valid_bits` and store `NaN` so charts draw gaps cleanly.
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

Cardinality caps (Brief §10.2) enforce bounded series usage:
- Top-50 clients by 24 h bytes + `other` bucket (`102` series)
- Top-120 firewall rules by hits + pinned rules (`240` series)
- Up to 32 WireGuard peers
- `ethtool -S` history only for pinned or well-known error counters (`≤ 32` per interface)

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
1. Collectors update T0 (1 s in-process ring) and roll up into T1 (10 s `tmpfs` ring) with zero flash I/O.
2. Every 60 s, the Rollup Engine builds:
   - 1 × T2 slot (`12 KiB` full / `8 KiB` lean) at index `(unix_sec / 60) % 20160`
   - Every 5 min: 1 × T3 slot at `(unix_sec / 300) % 25920`
   - Every 60 min: 1 × T4 slot at `(unix_sec / 3600) % 17520`
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

*(In the **lean profile** with 8 KiB slots, payload drops by one-third to `~14.5 MiB/day`.)*

### 4.3 Startup & Crash Recovery (Zero Replay Needed)
1. Read Superblock A and B; select the one with valid `header_crc32c` and highest `generation`.
2. Load active Series Catalog (`128 KiB`, verified by `catalog_crc32c`).
3. For each ring (T2, T3, T4, Event Ring), locate the newest valid slot head via **O(log N) binary search** on `(slot_seq, crc32c)` (~15 `pread`s per ring, `< 2 ms` total on eMMC). Any single torn slot at the head fails CRC and is cleanly skipped/overwritten on the next tick.
4. If `CLEAN_SHUTDOWN` and `T1_SNAPSHOT_VALID` are set and snapshot age `< 48 h`, decompress Region 7 into `/run/vyview/hot.ring` to restore 10 s resolution across planned reboots.

---

## 5. Query Path & Downsampling

1. **Tier selection** based on requested window `[from, to]` (Brief §7.5):
   - `≤ 15 min` → T0 (1 s in-process RAM)
   - `≤ 48 h` → T1 (10 s `tmpfs` RAM; falls back seamlessly to T2 across reboot gaps)
   - `≤ 14 d` → T2 (60 s in `vyview.ring`)
   - `≤ 90 d` → T3 (300 s in `vyview.ring`)
   - `> 90 d` (up to `2 y`) → T4 (3600 s in `vyview.ring`)
2. **Fast range reads:** Because slots are contiguous in time, reading a 24 h window from T2 (`1,440` slots) is a single sequential scan of `~17 MiB` (or wrap-around split into 2 sequential reads) that extracts all requested series simultaneously in one pass.
3. **Max-min preserving downsampling (M4 / LTTB):** Server caps output to `~1,500` points per series so uPlot renders in `< 1 ms` and SSE payloads remain tiny.
4. **Anomaly bands (Brief §7.13):** Computed server-side over the same hour-of-week buckets in T3/T4 using robust median + MAD (`baseline ± kσ`).

---

## 6. Upgrades, Persistence & Self-Monitoring

- **Recommended path on VyOS:** `/var/lib/vyview/vyview.ring` mounted on a dedicated eMMC partition or LV (survives both reboots and VyOS image upgrades without bloating `/config` backups).
- **Schema upgrades:** On a `format_version` bump, `vyview` runs an offline/startup in-place or streaming migration tool (or starts a fresh `vyview.v2.ring` while mounting the prior file read-only for seamless historical queries until it ages out).
- **Self-measurement:** `vyview` reads `/proc/self/io` (`write_bytes`) and exposes real write rate, ring fill/oldest timestamp, cardinality cap saturation, and eMMC `EXT_CSD` life-time estimates on `System › Storage` and `Settings › Retention`.
