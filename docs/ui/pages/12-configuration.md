# Configuration

> **Primary Question:** *"How is the router configured right now, what changed in recent commits, and how does this map to familiar CLI syntax?"* · **Routes:** `/config`, `/config/history`

---

## Layout & Key Views

Server-redacted configuration browser (config-as-code, strictly read-only in the web GUI; writes go via CLI or opt-in MCP `commit-confirm`):

- **Collapsible Config Tree:** Real expandable/collapsible hierarchical tree matching `show configuration` (never a flat "Records" table), where live entity nodes link straight to their operational status pages (e.g., `interfaces ethernet eth0` → `Interfaces › eth0`).
- **Set Commands (`Config as Code`):** Renders the full `set`-command list as **one copyable code block** (never one table row per command) with search/subtree filtering, JSON export tab, offline explain metadata, and **multi-vendor CLI cheat sheets** displaying equivalent commands for Juniper, Cisco, FortiGate, MikroTik, Hirschmann, and Siemens SCALANCE alongside VyOS.
- **Two-Revision Diff:** Pick any two archived revisions and inspect a **red/green unified or side-by-side diff** matching VyOS `compare`. Changed secrets display strictly as `•••••• (changed)` using per-boot keyed HMAC comparison, with **no reveal control** and **no Copy button on `[REDACTED]` fields**.
- **Audited Commit History:** Displays **who, how, and why** for every revision: user, change source (`CLI` / `API` / `MCP`), commit comment, whether **`commit-confirm`** was used, and **absolute date + time + timezone** (never ambiguous relative-only timestamps).

---

## Data Sources & Cadence

- **Running config & archive:** `inotify` watcher on `/config/archive/config.boot.N.gz` + on-commit `show configuration json` (`vyattacfg-ro` read-only group).
- **Redaction:** Server-side across all 19 secret config path classes *before* indexing, search, rendering, or diffing (see [Security & MCP §4](../../architecture/security-and-mcp.md#4-server-side-secret-redaction--privacy)).

---

## UX & Consistency Constraints

- Never expose edit, commit, rollback, or secret-reveal controls in the web GUI.
- Original field table: [brief-extracted.txt](../../reference/brief-extracted.txt) (p. 22).
