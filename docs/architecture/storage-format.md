# Single-File Multi-Ring Datastore (`vyview.ring`)

> Binary specification for vyview's pure-Go, pre-allocated, 4 KiB page-aligned multi-ring datastore (`vyview.ring`) and in-RAM `200 ms` / `5 s` hot tiers.

---

## 1. Design Goals & Constraints

1. **One active persistent file on disk (plus optional passive shadow copy):** All persistent time-series rings (`T2`/`T3`/`T4`), series metadata, curated events, alert transitions, diagnostic probe history, and an optional `T1` clean-shutdown snapshot live in **one fixed-size binary file** (`vyview.ring`, defaulting to `/var/lib/vyview/vyview.ring` or any mounted path via `--data-path` / `retention.path`, such as an external USB drive).
   - *Why one fixed file rather than per-flush tick/tock copies:* Modern eMMC FTLs wear-level logical block overwrites across physical erase blocks inside a fixed file, whereas duplicating a 160–778 MiB file per flush would destroy flash endurance. Instead, **A/B ping-pong superblocks** + **per-slot CRC32C** isolate any power-cut torn write to a single 12 KiB slot. Operators wanting an offline/exportable redundant copy can configure `--snapshot-path` to refresh a passive read-only backup on a slow schedule (e.g., daily) or clean shutdown.
2. **Fluid `200 ms` (5 Hz) live telemetry decoupled from flash I/O:** Fast kernel counters sample every `200 ms` into an in-process RAM ring (`T0`) and cascade with exact integer multipliers (`25× → 12× → 5× → 12×`) into `T1` (`5 s`, RAM/`tmpfs`) and the persistent flash tiers (`T2`–`T4`). Slow metrics (`1 s`+) piggyback on every 5th tick.
3. **Zero growth, zero fragmentation, zero compaction:** Pre-allocated once with `fallocate(2)`; file size never changes after initialization.
4. **Minimal eMMC/SD flash wear (`< 50 MB/day` host writes):** Time-major 4 KiB page-aligned slots, batched `pwrite(2)` + one `fdatasync(2)` every `60 s`, and self-describing slots (`timestamp`, `slot_seq`, `valid_bits`, `crc32c`) that eliminate routine superblock rewrites.
5. **Bounded memory (`< 40 MB` RSS):** Uses explicit `pread(2)`/`pwrite(2)` with pooled buffers instead of `mmap(2)`, preventing unbounded page-cache pressure and eliminating `SIGBUS` crash risk on I/O errors.
6. **Zero-migration series headroom:** Fixed capacity of `760` series (vs. `692` reference inventory) lets new interfaces, WireGuard peers, or pinned firewall rules claim free slots without rewriting the file.

---

## 2. Memory Tiers vs. Persistent File Regions

Every tier is an exact integer multiple of the tier below (`200 ms × 25 = 5 s`, `5 s × 12 = 60 s`, `60 s × 5 = 300 s`, `300 s × 12 = 3600 s`), ensuring zero resampling drift and clean UTC midnight alignment:

| Tier | Step ($\Delta t$) | Fan-In | Slots ($N$) | Span / Windows Served | Series Class | Location | Size |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **T0 Live** | `200 ms` (5 Hz) | Base tick | `2,048` | `6 m 50 s` (live 2 m / 5 m at 60 fps) | Core (`≤ 192`, `float32`) | Process RAM (contiguous 4 KiB slab) | `~1.5 MiB` |
| **T1 Hot** | `5 s` | `25 × T0` | `17,280` | `24 h` (15 m / 1 h / 6 h / 24 h) | Core (`≤ 192`, `min/avg/max/p95`) | RAM / `tmpfs` (`/run/vyview/hot.ring`) | `~50 MiB` *(in `tmpfs`)* |
| **T2 Warm** | `60 s` | `12 × T1` | `20,160` | `14 d` (1 h – 14 d) | All (`≤ 760`) | `vyview.ring` (Region 3) | `236.25 MiB` full / `157.50 MiB` lean |
| **T3 Cool** | `300 s` | `5 × T2` | `25,920` | `90 d` (30 d / 90 d) | All (`≤ 760`) | `vyview.ring` (Region 4) | `303.75 MiB` full / `202.50 MiB` lean |
| **T4 Cold** | `3600 s` | `12 × T3` | `17,520` | `2 y` (1 y / 2 y + anomaly bands) | All (`≤ 760`) | `vyview.ring` (Region 5) | `205.31 MiB` full / `136.88 MiB` lean |
| **Events / Alerts / Probes** | Event | — | `4,096` frames | `30 d` events / `90 d` alerts & probes | Structured records | `vyview.ring` (Region 6) | `16.00 MiB` |
| **T1 Shutdown Snapshot** *(opt.)* | On stop | — | 1 zstd blob | Latest `24 h` hot tier across planned reboots | T1 snapshot on `SIGTERM` | `vyview.ring` (Region 7) | `16.00 MiB` |

### Storage Profiles & Retention Bounds (`--profile` × `--retention`)

| Profile | Geometry & Aggregates | Total File Size | Est. Host Writes |
| --- | --- | --- | --- |
| **Full (`1`)** | `min/avg/max/p95` for all `760` series across `T2`/`T3`/`T4` | `~778 MiB` | `~28–32 MB/day` |
| **Lean (`2`)** | `min/avg/max/p95` for `192` core; `avg/max` for `568` extended | `~529 MiB` | `~19–22 MB/day` |
| **Compact (`3`)** | `T0` + `T1` (`5 s × 2,048` = `2 h 51 m`, `~6 MiB` RAM); omits `60 s` disk ring; persists `5 min × 2,048` (`7.1 d`) + `1 h × 8,784` (`366 d`) | `~160 MiB` full / `~117 MiB` lean | `< 5 MB/day` |
| **Live-only** | `T0` + `T1` in RAM/`tmpfs` only; zero persistent disk file (also automatic fallback if storage path is absent/full) | `0 B` | `0 B/day` |

- **Composable retention horizons:** Pair any profile with `max_retention` (`7d`, `90d`, `180d`, `1y`, `2y`). Short horizons omit cooler rings completely. Hard upper bounds reject configs that leave `< 20%` free partition space (protecting VyOS image upgrades) or fail a quick startup write-latency check on USB storage.
- **Startup health advisories (surfaced in Health & Alerts):**
  - `WARN` if storage target is absent/unmounted (graceful fallback to `live-only`).
  - `WARN` if insufficient free space exists for requested profile + retention.
  - `WARN` (early advisory) when eMMC wear exceeds `50%` (`EXT_CSD >= 0x06`), escalating at `≥ 0x08` (`70–80%+` worn) or `Pre-EOL >= 0x02`.

### In-RAM `T0` Slab & Fast/Slow Loop Split
- **Contiguous page-aligned slab:** `T0` holds `2,048 × float32 = 8,192 B` (`2 × 4 KiB` pages) per core series with power-of-two index masking (`idx & 2047`), allocated as one `1.5 MiB` contiguous backing slice (`192 × 8 KiB`) to eliminate heap fragmentation and reduce TLB misses on Cortex-A72.
- **Fast loop (`200 ms` / 5 Hz):** Samples interface `IFLA_STATS64`, `/proc/stat`, `/proc/softirqs`, and active queue drops. Matches DPAA1/FMan hardware shadow-counter coalescing (~100–250 ms) while eliminating jerky 1-second UI steps.
- **Optional `100 ms` (10 Hz) diagnostics burst:** When explicitly toggled on a LAN interface detail view (e.g., `iperf3`/`TRex` testing), live streaming may run at `100 ms` in-memory without changing `T1`–`T4` ring geometry.

---

## 3. Binary File Layout

All integers are **little-endian**. Every region starts on a **4 KiB (`4096 B`) boundary**:

```text
Offset (4 KiB aligned)
┌─────────────────────────────────────────────────────────────────────────┐
│ Region 0: Superblock A (4 KiB) + Superblock B (4 KiB)      [8 KiB]      │
│   Magic "VYVW", format_version, profile, generation++, CRC32C,          │
│   region table (offsets, slot sizes, capacities), T1 snapshot metadata  │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 1: Series Catalog A (128 KiB) + Backup B (128 KiB)  [256 KiB]    │
│   1,024 fixed 128-byte descriptors: series_id, canonical key hash+name, │
│   unit, metric kind (gauge/counter), tier class, active/retired flags   │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 2: Reserved / Padding to 1 MiB boundary             [760 KiB]    │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 3: Ring T2 — Warm  (60 s × 20,160 slots = 14 d)     [236.25 MiB] │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 4: Ring T3 — Cool  (300 s × 25,920 slots = 90 d)    [303.75 MiB] │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 5: Ring T4 — Cold  (3600 s × 17,520 slots = 2 y)    [205.31 MiB] │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 6: Circular Event / Alert / Probe Ring              [16.00 MiB]  │
│   4,096 × 4 KiB frames of length-prefixed CBOR/binary records           │
├─────────────────────────────────────────────────────────────────────────┤
│ Region 7: Optional T1 Hot-Ring Clean-Shutdown Snapshot     [16.00 MiB]  │
│   Zstd-compressed T1 ring written only on SIGTERM / ExecStop            │
└─────────────────────────────────────────────────────────────────────────┘
```

### 3.1 Superblock Structure (`4 KiB`, Ping-Pong A/B)

Written **only** on file creation, series catalog mutation, or clean shutdown (never on routine 60 s flushes):

| Field | Size | Description |
| --- | --- | --- |
| `magic` | `4 B` | ASCII `"VYVW"` |
| `format_version` | `4 B` | Schema version (starts at `1`) |
| `generation` | `8 B` | Monotonically increasing counter; highest valid CRC wins on mount |
| `profile` | `4 B` | `1 = full`, `2 = lean`, `3 = compact` |
| `max_series` / `core_series_cap` | `4 B` + `4 B` | `760` total / `192` core |
| `flags` | `4 B` | Bit 0: `CLEAN_SHUTDOWN`; Bit 1: `T1_SNAPSHOT_VALID` |
| `created_unix_ns` / `last_shutdown_unix_ns` | `8 B` + `8 B` | Creation and last clean-shutdown timestamps |
| `regions[8]` | `8 × 48 B` | Per-region `kind`, `offset`, `length`, `slot_bytes`, `slot_count`, `step_sec` |
| `catalog_crc32c` | `4 B` | Castagnoli CRC32C of active Series Catalog |
| `t1_snapshot_len` / `crc` | `12 B` | Compressed length and CRC32C of Region 7 snapshot |
| `reserved` + `header_crc32c` | padded + `4 B` | Zero-filled to byte `4092`, then Castagnoli CRC32C over `[0..4092)` |

### 3.2 Time-Major Slot Block Geometry

Every slot in `T2`, `T3`, and `T4` is atomic, page-aligned, and self-verifying:

- **Full profile slot (`12,288 B` = exact `3 × 4 KiB` pages):**
  - `128 B` header: `slot_magic` (`0x5653`), `flags`, `crc32c`, `unix_sec` (aligned bucket start), `slot_seq` (monotonic per ring), `valid_bits` (`96 B` = `768` bits), `reserved` (`8 B`).
  - `760 × 16 B = 12,160 B` aggregates: `[min, avg, max, p95]` as `4 × float32`.
- **Lean profile slot (`8,192 B` = exact `2 × 4 KiB` pages):**
  - `128 B` header + `192 × 16 B` core (`[min, avg, max, p95]`) + `568 × 8 B` extended (`[avg, max]`) + `448 B` zero padding.

### 3.3 Rate Math, Gaps & Peak Preservation
- **Monotonic $\Delta t$ rates:** At `200 ms` cadence, even `5–10 ms` of scheduler jitter introduces a `2.5–5%` rate error if dividing by a constant interval. Every counter read is stamped with Go's monotonic clock and converted via:
  $$\text{Rate} = \frac{\Delta \text{counter}}{\Delta t_{\text{monotonic, sec}}}$$
  with 32/64-bit wrap, interface-reset, and ASK2 driver-reload reset protection applied before aggregation.
- **Gap-safe accumulators (`validCount`):** Active rollup buckets maintain `{validCount, sum, min, max, histogram}` in RAM. When a bucket closes, `avg = sum / validCount`. If `validCount == 0` (link down, collector timeout, reboot), `valid_bits` is cleared and `NaN` is stored so charts render honest gaps instead of synthetic zeros.
- **Peak preservation:** Every tier from `T1` upward persists `max` alongside `avg` so `uPlot` can plot the sustained average with a translucent peak envelope behind it.

### 3.4 Series Catalog & Cardinality Caps

Region 1 stores double-buffered catalogs of `1,024 × 128 B` descriptors (`series_id`, `class`, `kind`, `unit`, `state` `free|active|retired`, `key_hash`, `first_seen_sec`, `last_seen_sec`, `96 B` UTF-8 canonical key such as `iface:eth0:rx_bps`).

Strict cardinality caps keep RAM and disk bounded under `8k–50k+` concurrent flows:
- **Clients:** Top-50 by 24 h bytes + `other` bucket (`102` series), tracked via a bounded **Space-Saving / Heavy-Hitter Top-K sketch**.
- **Firewall rules:** Top-120 by hits + pinned rules (`240` series).
- **WireGuard peers:** Up to `32` peers (including `wgfp_rx` fast-path counters when ASK2 is active).
- **Offload & hardware:** Per-port `offload_ratio`, aggregate HW/SW flows, CAAM job rates, and `≤ 32` pinned/well-known `ethtool -S` MAC/FIFO error counters per interface.
- Cap saturation is surfaced explicitly in the UI rather than silently dropping entities.

### 3.5 Circular Event, Alert & Probe Ring (Region 6, `16 MiB`)

Divided into `4,096 × 4 KiB` frames (`frame_magic`, `frame_seq`, `unix_sec`, `used_bytes`, `crc32c`, packed CBOR/binary records) holding curated events (`30 d` timeline & chart annotations), alert transitions (`90 d`), and scheduled diagnostic probe results (`30 d` `mtr` hop diffs, `90 d` bufferbloat/speed tests). Flushed at most once every `5 s` when dirty, or coalesced into the `60 s` `fdatasync`.

---

## 4. Write Path, Flash Wear & Crash Recovery

### 4.1 Statistical Cross-Tier Rollups (Every `60 s`)
Every `60 s` (`12 × T1`), the rollup engine writes one `T2` slot (plus one `T3` slot every `5 min` and one `T4` slot every `60 min`) via `pwrite(2)` followed by a single `fdatasync(fd)`.

Cross-tier rollups (`T1 → T2 → T3 → T4`) aggregate **statistically** rather than naively subsampling:
- `min = min(child.min)`
- `max = max(child.max)`
- `avg = sum(child.avg × child.validCount) / sum(child.validCount)`
- `p95` merged from child 64-bin log histograms (or conservative `max(child.p95)` fallback)

### 4.2 Daily Flash-Wear Budget (Full Profile, 760 Series)

| Operation | Writes / Day | Size / Write | Payload / Day |
| --- | --- | --- | --- |
| `T2` Warm (`60 s`) | `1,440` | `12 KiB` | `16.88 MiB` |
| `T3` Cool (`300 s`) | `288` | `12 KiB` | `3.38 MiB` |
| `T4` Cold (`3600 s`) | `24` | `12 KiB` | `0.28 MiB` |
| Event / alert / probe frames | `~200` | `4 KiB` | `0.78 MiB` |
| **Total payload** | | | **`~21.3 MiB/day`** |
| **Est. host writes (incl. ext4 journal)** | | | **`~28–32 MB/day`** (budget: `< 50 MB/day`) |
| **Est. NAND writes (eMMC WAF 3×)** | | | **`~85–96 MB/day` (`~34 GB/year`)** |
| **10-year wear (16 GB MLC eMMC, ~46.9 TBW)** | | | **`~0.71%` of endurance** |

*(Lean profile drops payload to `~14.5 MiB/day`; Compact profile drops payload to `~4.4 MiB/day`. Compare against SQLite WAL at `~1,109 MB/day` host writes or naive series-major RRD at `~3,893 MB/day`.)*

### 4.3 Startup & Zero-Replay Crash Recovery
1. Read Superblock A and B; select the valid `header_crc32c` with the highest `generation`.
2. Load and verify the active `128 KiB` Series Catalog.
3. For each ring (`T2`, `T3`, `T4`, Events), locate the newest valid slot head via **O(log N) binary search** on `(slot_seq, crc32c)` (~15 `pread`s per ring, `< 2 ms` on eMMC). Any single torn head slot fails CRC and is cleanly skipped/overwritten on the next tick.
4. If `CLEAN_SHUTDOWN` and `T1_SNAPSHOT_VALID` are set and snapshot age `< 24 h`, decompress Region 7 into `/run/vyview/hot.ring` to preserve `5 s` resolution across planned reboots.

---

## 5. Query Path, Tier Warming & Downsampling

1. **Automatic tier selection:**
   - `≤ 6 min` (Live 2 m / 5 m) → **T0** (`200 ms`, RAM)
   - `≤ 24 h` → **T1** (`5 s`, `tmpfs`; seamlessly falls back to `T2` across reboot gaps or in compact profile beyond `2 h 51 m`)
   - `≤ 14 d` → **T2** (`60 s`, or `300 s` up to `7 d` in compact profile)
   - `≤ 90 d` → **T3** (`300 s`)
   - `> 90 d` (up to `2 y`) → **T4** (`3600 s`)
2. **Bounded on-disk tier warming:** Sequential reads extract all requested series in one pass. Extracted per-series slices from `T2`–`T4` are cached in a **bounded, TTL-evicted in-RAM query cache** (`≤ 4 MiB` total cap, `5 min` idle TTL) so repeated chart zooms/pans stay sub-millisecond without threatening the `< 40 MB` RSS budget.
3. **Downsampling & anomaly bands:** Server caps chart series to `~1,000–1,500` points using min/max-preserving M4/LTTB downsampling. Hour-of-week anomaly bands (`baseline ± kσ`) are computed server-side from `T3`/`T4` using robust median + MAD.

---

## 6. Schema Upgrades & Self-Monitoring

- **Schema upgrades:** On a `format_version` bump, vyview creates a fresh `vyview.vN.ring` and mounts the previous ring file **read-only** for historical reads until it ages out — avoiding risky in-place migrations on flash.
- **Self-monitoring (`System › Storage & Images` and `Settings`):** Exposes real host write rate from `/proc/self/io` (`write_bytes`), **health of both Superblock A and Superblock B**, **torn-slot count**, **saved vs. expected nominal slots** per tier, configured storage target (`eMMC` vs. USB `--data-path`), and kernel-sysfs eMMC `EXT_CSD` wear (`life_time` Type A/B + `pre_eol_info`).
