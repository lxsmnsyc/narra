# Log: field-shorthand-sweep

- Base: feat/batch-15 at 034ef48f. Compiler built with `sh selfhost/bootstrap.sh <scratch>/base` (release).
- Proof first: a scan for `name: name` fields (struct literal fields whose value is the identifier of the same name, comments skipped) over `selfhost/src` and `std/src` finds 563 sites on 034ef48f.
- Sweep: one regex pass, `(?<![\w.:"])(\w+): \1(?=\s*(?:[,})\]]|$))` replaced by `\1`, over `selfhost/src/*.nra` and `std/src/*.nra`. 563 sites in 73 files, then `narrac fmt` (only lexer.nra changed layout: one call fits on one line again).
- After the sweep the same scan finds 0 sites.
- `sh scripts/test.sh --suite generated --update` rewrote `stdtext.nra` (it embeds the std sources).
- Emit comparison, same compiler (base): `narrac emit selfhost` of the unswept tree with the swept `stdtext.nra` against the swept tree. Identical except 26 `src/lexer.nra:LINE:COL` position strings, which shift by 4 lines where fmt joined the call; identical once those strings are masked.
- Seed regenerated twice with `sh scripts/seed-update.sh`.
