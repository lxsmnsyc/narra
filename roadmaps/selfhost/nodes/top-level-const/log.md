# Log: top-level-const

- Design: a constant is checked lazily at first reference (cycle = reference to a busy constant). Scalars are inlined at use; others get a getter that builds on first use. `narra_consts_release` frees them before the leak report.
- Codes taken: E0190-E0193 (declarations range, past opaque-distinct's likely E0120+), E0290 (expressions range). Expect renumbering on rebase.
- Gotcha: do not reset `c.closure` while checking a constant; function types registered there would be lost.
- Fast check passes.
