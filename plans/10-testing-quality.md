# Testing and quality (§14; PDF p. 65)

| Area | Approach and gate |
|---|---|
| Parsers | Golden captures for every command from VyOS 1.4 LTS, 1.5, rolling, x86 and reference arm64; CI every PR, rolling captures weekly |
| Robustness | Go fuzz text parsers and signal decoder, golden corpus; no panics; 10 min/target nightly |
| UI | Fake collectors replay/synthesize link flap, BGP down, pool full, 50k flows and offload full |
| Visual | Playwright every page/theme/breakpoint; review pixel diffs |
| Accessibility | axe, keyboard journeys and contrast; zero serious/critical |
| Read-only | Route, exec/API allowlist and netlink type tests; static check `os/exec` only in runner; security review |
| Security | CSP headers, static Datastar expressions lint, every secret redaction, govulncheck |
| Load | 50k conntrack, 200 clients, four browsers; within §12 budgets ×2 |
| Soak | 72 h reconnect storms, hide/show, minute range changes; no leak/stuck streams; RSS slope <1 MB/day |
| Performance | Renderer, rollup and search benchmarks; no >10% regression; payload <150 KB |
| Upgrade | Migration tests and hardware image upgrade; preserve history according to chosen persistence |

Authentication POST exceptions are explicitly described in §11.5 and Appendix E; distinguish them from read-only data-plane routes when implementing route tests.
