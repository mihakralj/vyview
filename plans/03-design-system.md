# Visual design system (§6; PDF pp. 25–27)

Dark-first operational UI; CSS variables are the source of truth and UnoCSS theme keys reference them. Use color with icon shape and label, never alone.

| Token | Dark | Light | Use |
|---|---|---|---|
| bg | #0B1220 | #F7F9FC | Page |
| rail | #0D1526 | #EEF2F7 | Navigation |
| surface | #111A2E | #FFFFFF | Cards |
| surface-2 | #16213A | #F1F5F9 | Inputs, headers |
| border | #22304D | #D5DEE9 | Dividers |
| text | #E6EDF7 | #0F172A | Primary |
| text-muted | #93A1B8 | #475569 | Secondary |
| accent | #22D3EE | #0E7490 | Focus, links |
| accent-2 | #14B8A6 | #0F766E | Secondary |
| ok | #10B981 | #047857 | Healthy |
| warn | #F5A524 | #B45309 | Degraded |
| crit | #F0573A | #B91C1C | Failed |
| info | #60A5FA | #1D4ED8 | Informational |
| unknown | #8B95A5 | #64748B | No data |

Chart slots, dark: `#22D3EE #A78BFA #F59E0B #34D399 #F472B6 #60A5FA #FB7185 #FACC15`; light: `#0891B2 #7C3AED #D97706 #059669 #DB2777 #2563EB #E11D48 #CA8A04`. Stable entity colors with collision resolution; rx slot 1, tx slot 2, other grey.

IBM Plex Sans/system-ui: title 22/28 px 700, card title 13/18 600, body 13/20 (compact 12/16), KPI 22–28 px (wall 40), caption 11/14. IBM Plex Mono/ui-monospace 12/18 for copyable values; tabular numerals. The demo embeds the fonts locally; this replaces the brief's original Inter and JetBrains Mono recommendation.

4 px grid; spacing 4/8/12/16/20/24/32/48 px; row heights 40/30/48 px. The demo uses flat, square panels and controls with separators and active indicators, replacing the brief's rounded-corner recommendation. Phone <640: 4 columns; tablet 640–1023: 8; desktop 1024–1919: 12; wall ≥1920 or kiosk: 12 and larger type. Card, table, units and full icon mappings are in the [extraction](brief-extracted.txt) pp. 25–27 and appendix D.

Tables use sticky headers, virtual rows plus 10 overscan for large sets, server slices above 2,000 rows, persisted columns and structured validated filters. Rates use SI bits/s, traffic SI bytes (IEC toggle), memory IEC, latency ms/µs, optical power dBm. Motion is 150–200 ms, live updates never animate; reduce-motion and kiosk disable transitions/flashes.
