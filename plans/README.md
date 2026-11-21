# vyview plans

The authoritative source is the [design brief v0.1.0](VyOS-ViewOnly-Dashboard-Design-Brief.pdf). The [extracted text](brief-extracted.txt) preserves the complete 76-page extraction, including detailed tables and sketches; its page markers refer to PDF pages. It has extraction artifacts and does not reproduce the mockup images. The plans below organize its implementation content; consult the PDF when visual layout or exact formatting matters. ASK2 sources and mockup data are illustrative, not established interfaces. Resource figures are estimates pending measurements.

- [Review and initial steps](00-review-and-initial-steps.md): coordinator review, initial sequence, risks and decisions.
- [Overview](00-overview.md): goals, personas, metrics and principles (§1–3).
- [Information architecture](01-information-architecture.md): navigation, inventory and state (§4).
- [Page specifications](02-pages/README.md): all 18 page and cross-page specs (§5).
- [Design system](03-design-system.md): tokens, typography, layout, tables and motion (§6).
- [Dashboard features and mockups](04-dashboard-features.md): behaviors and mockup callouts (§7–8).
- [Architecture](05-architecture.md): collectors, read-only execution, SSE, charts and assets (§9).
- [Storage and retention](06-storage-retention.md): inventory, tiers, wear and persistence (§10).
- [Security](07-security.md): privilege boundaries, authentication, CSP, redaction and privacy (§11).
- [Performance budgets](08-performance-budgets.md): costs, limits and measurement (§12).
- [Packaging and deployment](09-packaging-deployment.md): deployment options, configuration and build (§13).
- [Testing and quality](10-testing-quality.md): test and release gates (§14).
- [Roadmap and risks](11-roadmap-risks.md): release checklists, risks and open decisions (§15).
- [Appendices](appendices.md): source catalogue, implementation sketches, icon and route maps, glossary and references (A–G).
