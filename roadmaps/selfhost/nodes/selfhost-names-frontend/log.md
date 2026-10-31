# Log: selfhost-names-frontend

- Baseline on main (54773bb): Bodies 143 compared / 202 skipped; subtest pass/skip counts saved per test.
- Exported names and AST/struct fields read by other modules keep their names (all exports are used by main, modules or checker files).
- lexer.nra renamed; counts same as main.
- parser.nra, astdump.nra, idcheck.nra, manifest.nra renamed one at a time; counts same as main after each.
- Narra rejects shadowing in pattern bindings (`block is Some(block)`); fixed in idcheck.
- Literal check: all string/char/template text identical to main.
- Fast check passed (internal/narra 201 s).
