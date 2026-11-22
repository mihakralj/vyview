# Settings

> **Primary Question:** *"How should vyview display themes, units, timezones, and what are its active thresholds, retention, and integrations?"* · **Route:** `/settings`

---

## Layout & Key Views

Preferences belonging strictly to vyview itself (never VyOS router configuration), organized into **separate labelled radio / segmented control groups** (WCAG 2.2 AA, consistent capitalization — never mixing independent settings inside one tab bar):

- **Appearance:** Theme (`Dark` / `Light` / `System`), Density (`Comfortable` / `Compact` / `Wall`), Reduced motion toggle.
- **Demo Scenario *(demo builds only)*:** Separate labelled control group (`Normal` / `Warning` / `Unavailable`).
- **Display & Formatting:**
  - **Units:** Rates in bits/s vs. bytes/s, traffic in SI vs. IEC, temperature in `°C` vs. `°F`
  - **Time Zone & 24-Hour Clock:** Drives unambiguous dated timestamps across Overview, Configuration, Logs & Events, and Diagnostics
  - **Defaults & Kiosk:** Default time range, live update cadence, hidden-tab stream suspension, kiosk rotation interval, and pinned views (JSON export/import)
- **Operational Visibility (Read-Only Summary):**
  - **Warning Thresholds** displayed in plain words (configured outside the GUI)
  - **Data Retention:** Active `vyview.ring` profile, tier spans, fill `%`, storage path (`eMMC` vs. USB), and Superblock A/B status
  - **Integrations Status:** Prometheus `/metrics` endpoint status (enabled/disabled) and **MCP access status** (disabled / read-only / opt-in write via `commit-confirm`)

---

## Data Sources & Cadence

- Client display preferences persist in browser `localStorage` (demo build carries scenario/theme in URL parameters).
- Operational summaries read once on page load from vyview runtime config and `vyview.ring` metadata.

---

## UX & Consistency Constraints

- Never imply that changing a setting on this page modifies VyOS router configuration.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 24).
