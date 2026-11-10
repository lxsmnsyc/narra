# Log: or-patterns

- Checked first: no PatOr, no or-pattern parsing on feat/batch-3. Built from scratch.
- New node `Pattern.PatOr(id, pos, alternatives)`. `parsePattern` reads `|` between single patterns; the bar may start the next line (for the long fmt layout). Nested lists work because every pattern position calls `parsePattern`.
- Checker: the first alternative declares the locals. Each later alternative resolves with its own `seen` map, is compared with the first (names E0967, types E0968), then its PatName nodes are pointed at the first alternative's locals. The locals the later alternative created stay in the body, marked used.
- `is` with alternatives anywhere in the pattern: E0974, before the pattern is resolved.
- Exhaustiveness: rows whose first column is an or-pattern are expanded into one row per alternative (`expandAlternativeHeads`). `isIrrefutable` of an or-pattern is true when one alternative is.
- Unreachable alternatives: the repo had no unreachable-arm check at all. Added `isUseful` (patterns.nra), conservative (literals of infinite types only count wildcard rows, so it never reports a reachable alternative). It runs only for arms that are or-patterns, so plain arms behave as before. The reserved codes are only E0967, E0968, E0974, so the unreachable message uses E0974, which `is` also uses (two rows; the orchestrator chose it).
- Emitter: the condition of an or-pattern is `(a) || (b)`; the locals are declared unset before an `if / else if / else` chain that assigns from the first matching alternative (`bindAlternatives`). `bindPattern` got an `assignOnly` flag for nested lists.
- fmt: `group(first, indent(line "| " next ...))`. Fits: one line. Does not fit: first alternative on the arm line, each further alternative on its own line starting with `|`, indented. The brief says "one alternative per line starting with `|`"; the first alternative has no bar because the arm starts there.
- Seed: no or-pattern used in selfhost/src, seed not regenerated.
- Docs: SPEC.md (Pattern matching), DECISIONS.md D98 status, compiler/README.md, guide chapter 5 (one `narra fragment`).
