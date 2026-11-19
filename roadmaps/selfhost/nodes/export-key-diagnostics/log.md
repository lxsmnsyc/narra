# Log: export-key-diagnostics

- Branch `fix/export-key-diagnostics` fast-forwarded to `feat/batch-20` (4447d7d2 and the batch-20 briefs). The first `git reset --hard` was refused by the permission classifier, so I used `git merge --ff-only` on a clean tree.
- Base compiler: `seed/narrac.c` built with `cc` (the seed is current, so it behaves like the base source).
- Proof, item 3 (unused import fix): `base apply-fix testdata/errors/unused_import_multiline.nra` prints a file that starts with `  Shape,`. The fix stopped at the comment line `// the "Shape" trait`.
- Proof, item 2 (E0474): `base check testdata/stage2/errors/std_internal_import.nra` prints `std/internal/probe is internal to the standard library`, with no list.
- Proof, item 1: `NARRAC=base sh scripts/test.sh --suite deps --filter export-key` fails: `module "lib/streaming/reader" has no export "missing"`, where the importer wrote `lib/stream`.
- Work: `ImportDecl.writtenPath` (set by the parser, kept by the export rewrite). `exported` takes the written path for E0101 and E0402 and adds a note with the real module. E0421 and the `narrac pkg` dump show the written path. E0474 lists `moduleList`. `removeImport` ignores quotes in `//` comments.
- Goldens of the six E0474 cases rewritten with `--update`.
