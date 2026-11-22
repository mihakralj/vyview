# System

> **Primary Question:** *"Is the router hardware — CPU, memory, thermals, flash endurance, boot images, and vyview itself — healthy?"* · **Route:** `/system`

---

## Layout & Key Views

- **Resources Tab:** Router hardware/model identity, uptime, per-core CPU heatmap (`200 ms` fast loop), load average and Linux PSI (`/proc/pressure/*`), memory breakdown, **TMP431 sensors** (explicitly labelled CPU, PCB, SFP+ `hwmon` temperatures and fan RPM), and top-10 processes by CPU/RSS (collected only while tab is open; strictly no kill/signal actions).
- **Storage & Images Tab (No Magic Numbers, No Generic Records Block):**
  - **eMMC Wear & Health:** Framed unambiguously as **"eMMC wear"** (`%` worn) and **"Remaining life"** with visual bars translating all three `EXT_CSD` registers into plain language:
    - **Wear Type A (MLC/SLC)** and **Wear Type B:** translate hex `0x01`–`0x0B` (e.g., `0x03` → `"20–30% worn (70–80% life remaining)"`), early advisory warning at `> 50%` worn (`≥ 0x06`), high warning at `≥ 0x08`.
    - **Pre-EOL Info (primary alarm):** `0x01` → `"Normal"`, `0x02` → `"Warning (80% of reserved blocks consumed)"` (red/critical), `0x03` → `"Urgent (90%+ reserved blocks consumed)"` (red/critical).
  - **Datastore Self-Health (`vyview.ring`):** Reports **health of both Superblock A and Superblock B**, **torn-slot count**, **saved vs. expected nominal slots** per tier, configured storage target (`eMMC` vs. USB `--data-path`), and real host write rate from `/proc/self/io` (`MB/day` vs. `< 50 MB/day` budget).
  - **Boot Images:** Lists **all installed VyOS images** (not just the active one), marking which is currently running, which boots next (`default`), version, and build date.
  - **Disk Usage:** Per-partition used/free bars (`/`, `/boot`, `/config`, `/var/lib/vyview`) with an explicit **"Room for new VyOS image"** indicator (warn `> 85%`).
- **Units, Kernel & IRQs Tabs:** systemd units (read-only), kernel version/cmdline/taint/modules, NIC drivers/firmware, per-CPU IRQ/`softirq`/`/proc/net/softnet_stat` distribution (surfacing RSS imbalance and backlog drops), and **vyview self-health** (collector tick duration per source, dropped ticks, RSS).

---

## Data Sources & Cadence

- **CPU & softirq:** `200 ms` (`/proc/stat`, `/proc/softirqs`).
- **Memory, PSI, thermals, IRQs & `softnet_stat`:** `1–5 s` (`/proc/meminfo`, `/proc/pressure/*`, `/sys/class/hwmon/*`, `/proc/net/softnet_stat`).
- **eMMC `EXT_CSD` (`1 h`):** Read directly from unprivileged Linux sysfs (`/sys/block/mmcblk*/device/life_time` and `pre_eol_info`) — **`mmc-utils` is not required**.
- **Processes (`5 s` while open) & units (`30 s`):** Direct `/proc/<pid>` reads (no `ps` fork) and D-Bus `ListUnits`.

---

## UX & Consistency Constraints

- Never offer process kill, systemd restart, reboot, or VyOS image-switch actions in the web GUI.
- Keep uptime on **Resources**, not duplicated inside individual boot-image rows.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (pp. 21–22).
