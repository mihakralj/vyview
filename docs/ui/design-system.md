# Visual Design System

> Defines vyview's color tokens, typography, iconography, layout geometry, table conventions, units, and motion rules.

---

## 1. Visual Philosophy

vyview is a **dark-first operational interface** designed for clarity under incident pressure and zero visual clutter. CSS custom properties (`tokens.css`) are the single source of truth; UnoCSS theme keys reference them directly so light/dark themes and layout densities switch without rebuilding CSS.

- **Never rely on color alone:** Every semantic status pairs color with a distinct icon shape and a clear text label (WCAG 2.2 AA).
- **Flat, square engineering geometry:** Panels, cards, inputs, and controls use crisp flat surfaces, hairline dividers, and explicit active indicators rather than decorative drop shadows or heavy rounded corners.
- **Purpose-built views:** Follow the five UX design principles in [Product Overview](../architecture/overview.md#3-five-binding-ux-design-principles) — lead with a plain-language summary, show one truth per metric, and translate every raw code or threshold into human terms.

---

## 2. Color Tokens

### Semantic & Surface Palette

| Token | Dark | Light | Role |
| --- | --- | --- | --- |
| `--bg` | `#0B1220` | `#F7F9FC` | Page canvas |
| `--rail` | `#0D1526` | `#EEF2F7` | Left activity rail |
| `--surface` | `#111A2E` | `#FFFFFF` | Cards and panels |
| `--surface-2` | `#16213A` | `#F1F5F9` | Inputs, table headers, hover fills |
| `--border` | `#22304D` | `#D5DEE9` | Hairline dividers and rules |
| `--text` | `#E6EDF7` | `#0F172A` | Primary text |
| `--text-muted` | `#93A1B8` | `#475569` | Secondary labels and captions |
| `--accent` | `#22D3EE` | `#0E7490` | Focus rings, active state, links |
| `--accent-2` | `#14B8A6` | `#0F766E` | Secondary accent |
| `--ok` | `#10B981` | `#047857` | Healthy / up / pass |
| `--warn` | `#F5A524` | `#B45309` | Degraded / warning / standby |
| `--crit` | `#F0573A` | `#B91C1C` | Failed / critical / alarm |
| `--info` | `#60A5FA` | `#1D4ED8` | Informational / event marker |
| `--unknown` | `#8B95A5` | `#64748B` | Unavailable / uncollected |

### Chart Series Palette (16 Visually Distinct Complementary Slots)

Sixteen categorical tokens (`--chart-1` through `--chart-16`) cover dense router charts — up to 12 interfaces, 8 CAKE tins, top-N client stacked areas, Sankey flows, and multi-zone topologies — without adjacent hue collisions. Every slot exceeds the WCAG 2.2 non-text graphic contrast threshold (`≥ 3.0:1`, with all dark tokens `≥ 5.8:1` against `#111A2E` and all light tokens `≥ 3.68:1` against `#FFFFFF`).

| Token | Hue Name | Dark (`#111A2E` surf.) | Light (`#FFFFFF` surf.) | Primary Usage & Complementary Pairing |
| --- | --- | --- | --- | --- |
| `--chart-1` | Cyan | `#22D3EE` | `#0891B2` | Series 1 · RX / Download · Hardware offload (`HW`) |
| `--chart-2` | Violet | `#A78BFA` | `#7C3AED` | Series 2 · TX / Upload (paired dashed with `--chart-1`) |
| `--chart-3` | Amber | `#F59E0B` | `#B45309` | Series 3 · Warm complement to Cyan |
| `--chart-4` | Emerald | `#34D399` | `#059669` | Series 4 · Cool-green complement to Violet/Pink |
| `--chart-5` | Pink | `#F472B6` | `#DB2777` | Series 5 · High-separation magenta-pink |
| `--chart-6` | Blue | `#60A5FA` | `#2563EB` | Series 6 · Distinct cobalt blue |
| `--chart-7` | Rose | `#FB7185` | `#E11D48` | Series 7 · Crimson-rose accent |
| `--chart-8` | Yellow | `#FACC15` | `#A16207` | Series 8 · High-luminance yellow / ochre-gold |
| `--chart-9` | Orange | `#FB923C` | `#C2410C` | Series 9 · Software slow-path (`SW`, complements Cyan) |
| `--chart-10` | Teal | `#2DD4BF` | `#0F766E` | Series 10 · Deep seafoam between Cyan and Emerald |
| `--chart-11` | Indigo | `#818CF8` | `#4338CA` | Series 11 · Deep blue-violet |
| `--chart-12` | Lime | `#A3E635` | `#4D7C0F` | Series 12 · Yellow-green (distinct from Emerald & Yellow) |
| `--chart-13` | Fuchsia | `#E879F9` | `#A21CAF` | Series 13 · Bright purple-magenta |
| `--chart-14` | Sky | `#38BDF8` | `#0369A1` | Series 14 · Cerulean azure |
| `--chart-15` | Coral | `#F87171` | `#B91C1C` | Series 15 · Warm vermilion-red |
| `--chart-16` | Ochre | `#D4A373` | `#92400E` | Series 16 · Warm sand / bronze |

**Series Assignment & Color-Blind Safety Rules:**
- Entities receive deterministic hash-assigned slots with collision resolution so the same interface, VLAN, or client keeps the same color across charts on a page.
- Receive (`RX` / Download) always takes `--chart-1`; transmit (`TX` / Upload) takes `--chart-2` with a dashed stroke (`6 4`).
- On multi-line overlays exceeding 8 simultaneous lines, slots `1–8` render solid strokes while slots `9–16` add subtle dash patterns (`6 3` / `2 3`) so series remain distinguishable under deuteranopia, protanopia, and grayscale printing.
- Area and firewall-zone shading derive translucent fills via `color-mix(in srgb, var(--chart-N) 14%, transparent)` paired with `1 px` `var(--chart-N)` borders.

### Functional & Structural Chart Tokens

| Token | Dark | Light | Role |
| --- | --- | --- | --- |
| `--chart-other` | `#8B95A5` | `#64748B` | Neutral `"Other"` aggregate bucket in Top-N / stacked charts |
| `--chart-hw` | `#22D3EE` | `#0891B2` | Hardware-offloaded fast path (Plane A) |
| `--chart-sw` | `#FB923C` | `#C2410C` | Software / CPU slow path (Plane B, warm complement to `--chart-hw`) |
| `--chart-grid` | `#22304D` | `#E2E8F0` | Chart gridlines |
| `--chart-axis` | `#93A1B8` | `#475569` | Axis labels and tick marks |
| `--chart-crosshair` | `#94A3B8` | `#334155` | Synchronized cursor crosshair hairline |
| `--chart-band` | `rgba(96, 165, 250, 0.14)` | `rgba(37, 99, 235, 0.10)` | Hour-of-week anomaly / baseline envelope (`median ± k·MAD`) |
| `--chart-peak` | `rgba(34, 211, 238, 0.18)` | `rgba(8, 145, 178, 0.14)` | Translucent `max` peak envelope behind sustained `avg` |

### Sequential Heatmap Scale (CPU Cores, CEETM Queues & Load)

| Token | Dark | Light | Utilization Band |
| --- | --- | --- | --- |
| `--heat-0` | `#16213A` | `#F1F5F9` | `0%` / Idle |
| `--heat-1` | `#164E63` | `#CFFAFE` | `1–20%` Low |
| `--heat-2` | `#0E7490` | `#67E8F9` | `21–45%` Moderate |
| `--heat-3` | `#10B981` | `#34D399` | `46–70%` Active / Healthy |
| `--heat-4` | `#F59E0B` | `#F59E0B` | `71–85%` Elevated / Warn |
| `--heat-5` | `#F0573A` | `#DC2626` | `> 85%` Saturated / Critical |

---

## 3. Typography

vyview embeds **IBM Plex Sans** (1.3.6) and **IBM Plex Mono** (2.5.0) locally as Latin WOFF2 subsets under the SIL Open Font License 1.1 (zero external font requests; optional `-tags systemfonts` omits WOFF2 files).

| Role | Font Family | Size / Line-Height | Weight & Style |
| --- | --- | --- | --- |
| Page title | IBM Plex Sans / `system-ui` | `22 / 28 px` | `700` |
| Card / section title | IBM Plex Sans | `13 / 18 px` | `600` |
| Body | IBM Plex Sans | `13 / 20 px` (`12 / 16 px` compact) | `400` |
| KPI value | IBM Plex Sans | `22–28 px` (`40 px` wall/kiosk) | `600`–`700`, `tabular-nums` |
| Caption / metadata | IBM Plex Sans | `11 / 14 px` | `400`–`500` |
| Identifiers, IPs, CLI, counters | IBM Plex Mono / `ui-monospace` | `12 / 18 px` | `400`–`500`, `tabular-nums` |

---

## 4. Iconography

Icons are delivered as a single build-time subsetted SVG `<symbol>` sprite (`~64` symbols) with zero runtime Font Awesome JavaScript or webfonts:
- **Private builds (`-tags fapro`):** Licensed **Font Awesome Pro 7.3.1 Sharp Light** (`fa-sharp fa-light`) across navigation, status badges, KPIs, and chrome.
- **Public / upstream VyOS builds (default):** **Font Awesome Free 7.x** fallback sprite mapping every Pro-only glyph 1:1 onto Free symbols using identical IDs, preserving required `NOTICE` attribution.
- Never mix Classic and Sharp icon families on the same surface.

---

## 5. Layout Grid, Density & Breakpoints

- **Base grid:** `4 px` scale (`4 / 8 / 12 / 16 / 20 / 24 / 32 / 48 px`).
- **Row densities:** Comfortable `40 px`, Compact `30 px`, Wall/Kiosk `48 px`.
- **Responsive columns:**
  - Phone (`< 640 px`): `4` columns
  - Tablet (`640–1023 px`): `8` columns
  - Desktop (`1024–1919 px`): `12` columns
  - Wall / Kiosk (`≥ 1920 px`): `12` columns with enlarged KPI and label typography

---

## 6. Tables, Units & Motion

- **Data tables:** Sticky headers, virtual scrolling (`+10` overscan rows) for large sets, server-side filtering/sorting above `2,000` rows, persisted column visibility, and structured validated filters. Omit unnecessary filter/sort toolbars on tiny static tables (e.g., 5-row SFP DOM telemetry).
- **Formatting & units:**
  - Rates: SI bits/s (`Mb/s`, `Gb/s`) with user byte toggle in Settings
  - Traffic volume: SI bytes (IEC toggle available)
  - Memory & storage: IEC (`MiB`, `GiB`)
  - Latency: `ms` / `µs`; optical power: `dBm`; temperatures: `°C` (with `°F` toggle)
  - Timestamps: Full date + time + timezone (`24 h` default, configurable) across Configuration, Logs, Diagnostics, and Events
- **Motion & accessibility:** Interactive transitions are `150–200 ms`. **Live telemetry updates never animate or flash** (avoiding visual noise on dashboards and wall displays). `prefers-reduced-motion` and kiosk mode disable all transitions.
